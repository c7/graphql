package graphql

import (
	"net/http"
	"testing"
)

func TestOptions(t *testing.T) {
	t.Run("WithHTTPClient", func(t *testing.T) {
		c := &Client{}

		WithHTTPClient(http.DefaultClient)(c)

		if c.httpClient == nil {
			t.Fatalf("expected httpClient to be set")
		}
	})

	t.Run("ImmediatelyCloseReqBody", func(t *testing.T) {
		c := &Client{}

		ImmediatelyCloseReqBody()(c)

		if !c.closeReq {
			t.Fatalf("expected closeReq to be true")
		}
	})
}
