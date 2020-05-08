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
)

// Client is a client for interacting with a GraphQL API.
type Client struct {
	endpoint   string
	httpClient *http.Client

	// closeReq will close the request body immediately allowing for reuse of client
	closeReq bool
}

// NewClient makes a new Client capable of making GraphQL requests.
func NewClient(endpoint string, options ...Option) *Client {
	c := &Client{endpoint: endpoint}

	for _, option := range options {
		option(c)
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

type graphResponse struct {
	Data   interface{}
	Errors Errors
}
