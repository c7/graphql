package graphql

import "testing"

func TestNewRequest(t *testing.T) {
	q := "query { test }"

	r := NewRequest(q)

	if got, want := r.Query(), q; got != want {
		t.Fatalf("r.Query() = %q, want %q", got, want)
	}

	if got, want := len(r.Vars()), 0; got != want {
		t.Fatalf("len(r.Vars()) = %d, want %d", got, want)
	}

	r.Var("foo", 123)
	r.Var("bar", true)

	if got, want := len(r.Vars()), 2; got != want {
		t.Fatalf("len(r.Vars()) = %d, want %d", got, want)
	}

	v := r.Vars()

	if got, want := v["foo"].(int), 123; got != want {
		t.Fatalf(`v["foo"].(int) = %d, want %d`, got, want)
	}
}
