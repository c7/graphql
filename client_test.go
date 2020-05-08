package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		hf := func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Header.Get("X-Test"), "header-value"; got != want {
				t.Fatalf(`r.Header.Get("Example-Header") = %q, want %q`, got, want)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"test": "value in response",
				},
			})
		}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		c := NewClient(ts.URL, WithHTTPClient(nil))

		if c.httpClient == nil {
			t.Fatalf("expected httpClient to be set")
		}

		req := NewRequest("query { test }")

		req.Header.Add("X-Test", "header-value")

		var resp struct {
			Test string
		}

		if err := c.Run(context.Background(), req, &resp); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.Test, "value in response"; got != want {
			t.Fatalf("resp.Test = %q, want %q", got, want)
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		hf := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		c := NewClient(ts.URL)

		req := NewRequest("")

		err := c.Run(context.Background(), req, nil)
		if err == nil {
			t.Fatalf("expected error")
		}

		if got, want := err.Error(), "graphql: server returned a non-200 status code: 418"; got != want {
			t.Fatalf("err.Error() = %q, want %q", got, want)
		}
	})

	t.Run("decoding response", func(t *testing.T) {
		hf := func(w http.ResponseWriter, r *http.Request) {}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		c := NewClient(ts.URL)

		req := NewRequest("")

		err := c.Run(context.Background(), req, nil)
		if err == nil {
			t.Fatalf("expected error")
		}

		if got, want := err.Error(), "decoding response: EOF"; got != want {
			t.Fatalf("err.Error() = %q, want %q", got, want)
		}
	})

	t.Run("GraphQL errors", func(t *testing.T) {
		hf := func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": "first"},
					{"message": "second"},
				},
			})
		}

		ts := httptest.NewServer(http.HandlerFunc(hf))
		defer ts.Close()

		c := NewClient(ts.URL)

		req := NewRequest("")

		err := c.Run(context.Background(), req, nil)
		if err == nil {
			t.Fatalf("expected error")
		}

		if got, want := err.Error(), "graphql: first; second"; got != want {
			t.Fatalf("err.Error() = %q, want %q", got, want)
		}
	})
}

func TestClientRun(t *testing.T) {
	t.Run("Done context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c := &Client{}

		err := c.Run(ctx, nil, nil)

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
