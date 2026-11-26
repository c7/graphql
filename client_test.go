package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
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
			json.NewEncoder(w).Encode(map[string]any{
				"errors": []map[string]any{
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

func TestClientRunResponseBody(t *testing.T) {
	// truncated responds with a Content-Length larger than the body that is
	// written, and then closes the connection.
	truncated := func(status int, partial string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			conn, buf, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("hijack: %v", err)
				return
			}
			defer conn.Close()

			fmt.Fprintf(buf, "HTTP/1.1 %d %s\r\nContent-Length: 1000\r\n\r\n%s",
				status, http.StatusText(status), partial)
			buf.Flush()
		}
	}

	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{
			name:    "null body",
			handler: func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "null") },
			want:    "",
		},
		{
			name:    "truncated mid-JSON",
			handler: truncated(http.StatusOK, `{"data":{"te`),
			want:    "reading body: unexpected EOF",
		},
		{
			name:    "truncated mid-JSON with non-200 status",
			handler: truncated(http.StatusInternalServerError, `{"data":{"te`),
			want:    "reading body: unexpected EOF",
		},
		{
			name:    "truncated after valid JSON",
			handler: truncated(http.StatusOK, `{"data":{"test":"value"}}`),
			want:    "reading body: unexpected EOF",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			var resp struct {
				Test string
			}

			err := NewClient(ts.URL).Run(t.Context(), NewRequest("query { test }"), &resp)

			var got string
			if err != nil {
				got = err.Error()
			}

			if got != tc.want {
				t.Fatalf("err = %q, want %q", got, tc.want)
			}

			if tc.want != "" && !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("errors.Is(err, io.ErrUnexpectedEOF) = false, want true")
			}
		})
	}
}
