package api

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// MakeURLValues converts a request struct into a url.Values map.
func MakeURLValues(v interface{}) (url.Values, error) {
	values := make(url.Values)

	valElem := reflect.ValueOf(v).Elem()
	typElem := reflect.TypeOf(v).Elem()

	for i := 0; i < typElem.NumField(); i++ {
		field := typElem.Field(i)
		tagParams := strings.Split(field.Tag.Get("url"), ",")
		if len(tagParams) == 0 {
			continue
		}
		urlTag := tagParams[0]
		omitEmpty := len(tagParams) == 2 && tagParams[1] == "omitempty"

		if urlTag == "" || urlTag == "-" {
			continue
		}

		fieldValue := valElem.Field(i)
		if omitEmpty && fieldValue.IsZero() {
			continue
		}

		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			values.Set(urlTag, fieldValue.Interface().(time.Time).UTC().Format(time.RFC3339))
			continue
		}

		stringer, ok := fieldValue.Interface().(fmt.Stringer)
		if ok {
			values.Set(urlTag, stringer.String())
			continue
		}

		k := fieldValue.Kind()
		var s string
		switch k {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64:
			s = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
			reflect.Uint64:
			s = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Float32:
			s = strconv.FormatFloat(fieldValue.Float(), 'f', 4, 32)
		case reflect.Float64:
			s = strconv.FormatFloat(fieldValue.Float(), 'f', 4, 64)
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.Uint8 {
				s = string(fieldValue.Bytes())
			}
		case reflect.String:
			s = fieldValue.String()
		case reflect.Bool:
			s = fmt.Sprintf("%v", fieldValue.Bool())
		}
		values.Set(urlTag, s)
	}

	return values, nil
}
