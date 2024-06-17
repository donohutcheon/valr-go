# Valr Golang API Client

Go API wrapper for the [valr exchange](https://www.valr.com)

Based on the [luno exchange API wrapper](https://github.com/luno/luno-go)

### Installation

```
go get github.com/donohutcheon/valr-go
```

### Authentication

Public and private API keys can be generated within your account at the [exchange](https://www.valr.com)

Create a `.env` file in the root of your project and add the following:

```.env
VA_KEY_ID=<api_key_public>
VA_SECRET=<api_key_secret>
```

### Example usage

Refer to the `examples` directory for examples on how to use the http and websocket client.

### License

This is a derived work from [github.com/i-norden/valr-go](https://github.com/i-norden/valr-go) which is licensed under the [MIT](https://github.com/i-norden/valr-go/blob/master/LICENSE) license.