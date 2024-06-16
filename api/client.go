package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultBaseURL is the default base URL for the Valr API.
	defaultBaseURL = "https://api.valr.com/v1"
	// defaultTimeout is the default timeout for requests made by the client.
	defaultTimeout = 10 * time.Second
)

var ErrTooManyRequests = errors.New("too many requests")

// Client is a Valr API client.
type Client struct {
	httpClient   *http.Client
	rateLimiter  Limiter
	baseURL      string
	apiKeyPub    string
	apiKeySecret string
	debug        bool
}

// NewClient creates a new Valr API client with the default base URL.
func NewClient() *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: defaultTimeout},
		rateLimiter: NewRateLimiter(),
		baseURL:     defaultBaseURL,
	}
}

// SetAuth provides the client with an API key and secret.
func (cl *Client) SetAuth(apiKeyID, apiKeySecret string) error {
	if apiKeyID == "" || apiKeySecret == "" {
		return errors.New("valr: no credentials provided")
	}
	cl.apiKeyPub = apiKeyID
	cl.apiKeySecret = apiKeySecret
	return nil
}

// SetHTTPClient sets the HTTP client that will be used for API calls.
func (cl *Client) SetHTTPClient(httpClient *http.Client) {
	cl.httpClient = httpClient
}

// SetTimeout sets the timeout for requests made by this client. Note: if you
// set a timeout and then call .SetHTTPClient(), the timeout in the new HTTP
// client will be used.
func (cl *Client) SetTimeout(timeout *time.Duration) {
	if timeout == nil {
		cl.httpClient.Timeout = defaultTimeout
	} else {
		cl.httpClient.Timeout = *timeout
	}
}

// SetBaseURL overrides the default base URL. For internal use.
func (cl *Client) SetBaseURL(baseURL string) {
	cl.baseURL = strings.TrimRight(baseURL, "/")
}

// SetDebug enables or disables debug mode. In debug mode, HTTP requests and
// responses will be logged.
func (cl *Client) SetDebug(debug bool) {
	cl.debug = debug
}

func (cl *Client) do(ctx context.Context, method, path string,
	req, res interface{}, auth bool) error {

	url := cl.baseURL + "/" + strings.TrimLeft(path, "/")

	if cl.debug {
		log.Printf("valr: Call: %s %s", method, path)
		log.Printf("valr: Request: %#v", req)
	}

	var reqBody []byte
	if req != nil {
		values, err := MakeURLValues(req)
		if err != nil {
			return err
		}
		tags := findTags(path)
		for _, tag := range tags {
			url = strings.Replace(url, "{"+tag+"}", values.Get(tag), -1)
			values.Del(tag)
		}
		for key := range values {
			if values.Get(key) == "" {
				values.Del(key)
			}
		}
		if method == http.MethodGet {
			if values.Encode() != "" {
				url = url + "?" + values.Encode()
			}
		} else {
			reqBody, err = json.Marshal(req)
			if err != nil {
				return err
			}
		}
	}
	if cl.debug {
		log.Printf("Request URL: %s", url)
		log.Printf("Request body: %s", string(reqBody))
	}

	httpReq, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	httpReq = httpReq.WithContext(ctx)

	if method != http.MethodGet {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	cl.rateLimiter.Wait(ctx)
	if auth {
		httpReq.Header.Set("X-VALR-API-KEY", cl.apiKeyPub)
		now := time.Now()
		timestampString := strconv.FormatInt(now.UnixNano()/1000000, 10)
		path := strings.Replace(url, "https://api.valr.com", "", -1)
		signature := SignRequest(cl.apiKeySecret, timestampString, method, path, reqBody)
		httpReq.Header.Set("X-VALR-SIGNATURE", signature)
		httpReq.Header.Set("X-VALR-TIMESTAMP", timestampString)
		if cl.debug {
			log.Printf("X-VALR-API-KEY: %s", cl.apiKeyPub)
			log.Printf("X-VALR-SIGNATURE: %s", signature)
			log.Printf("X-VALR-TIMESTAMP: %s", timestampString)
		}
	}

	httpRes, err := cl.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	resBody, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return err
	}
	if cl.debug {
		log.Printf("Response: %s", string(resBody))
	}

	if httpRes.StatusCode == 429 {
		return ErrTooManyRequests
	}

	if httpRes.StatusCode/100 != 2 {
		log.Printf("valr: Call: %s %s\nvalr: Request: %s\nvalr: Response: %s\n", method, path, string(reqBody), string(resBody))
		return fmt.Errorf("valr: error response (%d %s)",
			httpRes.StatusCode, http.StatusText(httpRes.StatusCode))
	}

	return json.Unmarshal(resBody, res)
}

// getProtocol takes a URL string and returns its protocol (scheme).
func getProtocol(rawurl string) (string, error) {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to parse url for protocol"))
	}

	return parsedURL.Scheme, nil
}

func GetAuthHeaders(rawurl string, method string, apiKeyPub, apiKeySecret string, reqBody []byte) (http.Header, error) {
	headers := http.Header{}

	headers.Set("X-VALR-API-KEY", apiKeyPub)
	now := time.Now()
	timestampString := strconv.FormatInt(now.UnixNano()/1000000, 10)
	scheme, err := getProtocol(rawurl)
	if err != nil {
		return nil, err
	}
	var path string
	if scheme == "wss" {
		path = strings.Replace(rawurl, "wss://api.valr.com", "", -1)
	} else if scheme == "https" {
		path = strings.Replace(rawurl, "https://api.valr.com", "", -1)
	} else {
		return nil, errors.New("unsupported protocol")
	}
	signature := SignRequest(apiKeySecret, timestampString, method, path, reqBody)
	headers.Set("X-VALR-SIGNATURE", signature)
	headers.Set("X-VALR-TIMESTAMP", timestampString)

	return headers, nil
}

func findTags(str string) []string {
	tags := make([]string, 0)
	firstTag, remainder := findTag(str)
	if firstTag != "" {
		tags = append(tags, firstTag)
	}
	for remainder != "" {
		var nextTag string
		nextTag, remainder = findTag(remainder)
		if nextTag != "" {
			tags = append(tags, nextTag)
		}
	}
	return tags
}

func findTag(str string) (string, string) {
	leftIndex := strings.Index(str, "{")
	if leftIndex > 0 {
		rightIndex := strings.Index(str[leftIndex:], "}")
		if rightIndex >= 0 {
			return str[leftIndex+1 : rightIndex+leftIndex], str[rightIndex+leftIndex:]
		}
	}
	return "", ""
}

func SignRequest(apiSecret string, timestampString, verb, path string, body []byte) string {
	// Create a new Keyed-Hash Message Authentication Code (HMAC) using SHA512 and API Secret
	mac := hmac.New(sha512.New, []byte(apiSecret))

	mac.Write([]byte(timestampString))
	mac.Write([]byte(strings.ToUpper(verb)))
	mac.Write([]byte(path))
	mac.Write(body)
	// Gets the byte hash from HMAC and converts it into a hex string
	return hex.EncodeToString(mac.Sum(nil))
}
