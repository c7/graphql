package graphql

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkClientRun(b *testing.B) {
	const body = `{"data":{"user":{"id":"123","name":"Test User","email":"test@example.com","tags":["a","b","c"]}}}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	ctx := b.Context()

	b.ReportAllocs()

	for b.Loop() {
		req := NewRequest("query ($id: ID!) { user(id: $id) { id name email tags } }",
			Variable("id", "123"),
			Header("Authorization", "Bearer token"),
		)

		var resp struct {
			User struct {
				ID    string
				Name  string
				Email string
				Tags  []string
			}
		}

		if err := c.Run(ctx, req, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewJSONRequestBody(b *testing.B) {
	req := NewRequest("query ($id: ID!) { user(id: $id) { id name email tags } }",
		Variable("id", "123"),
		Variable("limit", 10),
	)

	b.ReportAllocs()

	for b.Loop() {
		if _, err := newJSONRequestBody(req); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkErrorsError(b *testing.B) {
	errs := Errors{{Message: "first"}, {Message: "second"}, {Message: "third"}}

	b.ReportAllocs()

	for b.Loop() {
		_ = errs.Error()
	}
}
