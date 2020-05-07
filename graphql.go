// Package graphql is a GraphQL client
// with no third party dependencies.
//
// Initial version based on
// https://github.com/machinebox/graphql
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is a client for interacting with a GraphQL API.
type Client struct {
	endpoint   string
	httpClient *http.Client

	// closeReq will close the request body immediately allowing for reuse of client
	closeReq bool
}

// ClientOption are functions that are passed into NewClient to
// modify the behaviour of the Client.
type ClientOption func(*Client)

// WithHTTPClient specifies the underlying http.Client to use when
// making requests.
//  NewClient(endpoint, WithHTTPClient(specificHTTPClient))
func WithHTTPClient(httpclient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpclient
	}
}

//ImmediatelyCloseReqBody will close the req body immediately after each request body is ready
func ImmediatelyCloseReqBody() ClientOption {
	return func(client *Client) {
		client.closeReq = true
	}
}

// NewClient makes a new Client capable of making GraphQL requests.
func NewClient(endpoint string, opts ...ClientOption) *Client {
	c := &Client{endpoint: endpoint}

	for _, optionFunc := range opts {
		optionFunc(c)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c
}

// Run executes the query and unmarshals the response from the data field
// into the response object.
// Pass in a nil response object to skip response parsing.
// If the request fails or the server returns an error, the first error
// will be returned.
func (c *Client) Run(ctx context.Context, req *Request, resp interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return c.runWithJSON(ctx, req, resp)
}

func (c *Client) runWithJSON(ctx context.Context, req *Request, resp interface{}) error {
	r, err := c.newJSONRequest(req)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(r.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, res.Body); err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	gr := &graphResponse{
		Data: resp,
	}

	if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("graphql: server returned a non-200 status code: %v", res.StatusCode)
		}

		return fmt.Errorf("decoding response: %w", err)
	}

	if len(gr.Errors) > 0 {
		return gr.Errors
	}

	return nil
}

func (c *Client) newJSONRequest(req *Request) (*http.Request, error) {
	body, err := newJSONRequestBody(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(http.MethodPost, c.endpoint, body)
	if err != nil {
		return nil, err
	}

	r.Close = c.closeReq

	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	r.Header.Set("Accept", "application/json; charset=utf-8")

	for key, values := range req.Header {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}

	return r, nil
}

// Errors contains all the errors that were returned by the GraphQL server.
type Errors []Error

func (ee Errors) Error() string {
	if len(ee) == 0 {
		return "no errors"
	}

	errs := make([]string, len(ee))
	for i, e := range ee {
		errs[i] = e.Message
	}

	return "graphql: " + strings.Join(errs, "; ")
}

// An Error contains error information returned by the GraphQL server.
type Error struct {
	// Message contains the error message.
	Message string
	// Locations contains the locations in the GraphQL document that caused the
	// error if the error can be associated to a particular point in the
	// requested GraphQL document.
	Locations []Location
	// Path contains the key path of the response field which experienced the
	// error. This allows clients to identify whether a nil result is
	// intentional or caused by a runtime error.
	Path []interface{}
	// Extensions may contain additional fields set by the GraphQL service,
	// such as	an error code.
	Extensions map[string]interface{}
}

// A Location is a location in the GraphQL query that resulted in an error.
// The location may be returned as part of an error response.
type Location struct {
	Line   int
	Column int
}

func (e Error) Error() string {
	return "graphql: " + e.Message
}

type graphResponse struct {
	Data   interface{}
	Errors Errors
}

// Request is a GraphQL request.
type Request struct {
	q    string
	vars map[string]interface{}

	// Header represent any request headers that will be set
	// when the request is made.
	Header http.Header
}

// NewRequest makes a new Request with the specified string.
func NewRequest(q string) *Request {
	req := &Request{
		q:      q,
		Header: make(map[string][]string),
	}

	return req
}

// Var sets a variable.
func (req *Request) Var(key string, value interface{}) {
	if req.vars == nil {
		req.vars = make(map[string]interface{})
	}

	req.vars[key] = value
}

// Vars gets the variables for this Request.
func (req *Request) Vars() map[string]interface{} {
	return req.vars
}

// Query gets the query string of this request.
func (req *Request) Query() string {
	return req.q
}

func newJSONRequestBody(req *Request) (*bytes.Buffer, error) {
	var body bytes.Buffer

	v := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     req.q,
		Variables: req.vars,
	}

	if err := json.NewEncoder(&body).Encode(v); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}

	return &body, nil
}
