package graphql

import "net/http"

// Request is a GraphQL request.
type Request struct {
	q    string
	vars map[string]interface{}

	// Header represent any request headers that will be set
	// when the request is made.
	Header http.Header
}

// RequestOption is an option function for a request.
type RequestOption func(*Request)

// Variable sets the provided key and value as a variable on the request.
func Variable(key string, value any) RequestOption {
	return func(req *Request) {
		req.Var(key, value)
	}
}

// Header sets the provided key and value as a header on the request.
func Header(key, value string) RequestOption {
	return func(req *Request) {
		req.Header.Set(key, value)
	}
}

// NewRequest makes a new request with the specified query and options.
func NewRequest(q string, options ...RequestOption) *Request {
	req := &Request{
		q:      q,
		Header: make(map[string][]string),
	}

	for _, option := range options {
		option(req)
	}

	return req
}

// Var sets a variable on this request.
func (req *Request) Var(key string, value any) {
	if req.vars == nil {
		req.vars = make(map[string]any)
	}

	req.vars[key] = value
}

// Vars gets the variables for this request.
func (req *Request) Vars() map[string]any {
	return req.vars
}

// Query gets the query string of this request.
func (req *Request) Query() string {
	return req.q
}
