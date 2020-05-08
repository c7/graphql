package graphql

import "net/http"

// Option are functions that are passed into NewClient to
// modify the behaviour of the Client.
type Option func(*Client)

// WithHTTPClient specifies the underlying http.Client to use when
// making requests.
//  NewClient(endpoint, WithHTTPClient(specificHTTPClient))
func WithHTTPClient(httpclient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpclient
	}
}

//ImmediatelyCloseReqBody will close the req body immediately after each request body is ready
func ImmediatelyCloseReqBody() Option {
	return func(client *Client) {
		client.closeReq = true
	}
}
