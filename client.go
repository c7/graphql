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
func (c *Client) Run(ctx context.Context, req *Request, resp any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return c.runWithJSON(ctx, req, resp)
}

func (c *Client) runWithJSON(ctx context.Context, req *Request, resp any) error {
	r, err := c.newJSONRequest(ctx, req)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body := &readErrRecorder{r: res.Body}

	gr := graphResponse{
		Data: resp,
	}

	decodeErr := json.NewDecoder(body).Decode(&gr)

	// Drain any remaining body so the connection can be reused. Read errors
	// are recorded by body and take precedence over decoding errors.
	_, _ = io.Copy(io.Discard, body)

	if body.err != nil {
		return fmt.Errorf("reading body: %w", body.err)
	}

	if err := decodeErr; err != nil {
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

func (c *Client) newJSONRequest(ctx context.Context, req *Request) (*http.Request, error) {
	body, err := newJSONRequestBody(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, body)
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

func newJSONRequestBody(req *Request) (*bytes.Reader, error) {
	v := struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}{
		Query:     req.q,
		Variables: req.vars,
	}

	body, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}

	return bytes.NewReader(body), nil
}

// readErrRecorder wraps an io.Reader and records the first read error
// other than io.EOF, allowing transport errors to be told apart from
// decoding errors when decoding directly from a response body.
type readErrRecorder struct {
	r   io.Reader
	err error
}

func (rec *readErrRecorder) Read(p []byte) (int, error) {
	n, err := rec.r.Read(p)
	if err != nil && err != io.EOF && rec.err == nil {
		rec.err = err
	}

	return n, err
}

type graphResponse struct {
	Data   any
	Errors Errors
}
