package graphql

import "testing"

func TestNewRequest(t *testing.T) {
	q := "query { test }"

	r := NewRequest(q,
		Variable("qux", 123),
		Header("sit", "amet"),
	)

	if got, want := r.Query(), q; got != want {
		t.Fatalf("r.Query() = %q, want %q", got, want)
	}

	if got, want := len(r.Vars()), 1; got != want {
		t.Fatalf("len(r.Vars()) = %d, want %d", got, want)
	}

	if got, want := len(r.Header), 1; got != want {
		t.Fatalf("len(r.Header) = %d, want %d", got, want)
	}

	r.Var("foo", 123)
	r.Var("bar", true)

	if got, want := len(r.Vars()), 3; got != want {
		t.Fatalf("len(r.Vars()) = %d, want %d", got, want)
	}

	v := r.Vars()

	if got, want := v["foo"].(int), 123; got != want {
		t.Fatalf(`v["foo"].(int) = %d, want %d`, got, want)
	}
}
