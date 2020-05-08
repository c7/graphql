package graphql

import "testing"

func TestErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		errs := Errors{}

		if got, want := errs.Error(), "no errors"; got != want {
			t.Fatalf("errs.Error() = %q, want %q", got, want)
		}
	})

	t.Run("two errors", func(t *testing.T) {
		errs := Errors{
			{Message: "first"},
			{Message: "second"},
		}

		if got, want := errs.Error(), "graphql: first; second"; got != want {
			t.Fatalf("errs.Error() = %q, want %q", got, want)
		}
	})
}

func TestError(t *testing.T) {
	err := Error{Message: "test error"}

	if got, want := err.Error(), "graphql: test error"; got != want {
		t.Fatalf("err.Error() = %q, want %q", got, want)
	}
}
