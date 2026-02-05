package data

import (
	"testing"
)

func TestRemoveQuotedText(t *testing.T) {
	result := removeQuotedText(`Who's There?

On 2006-01-02 MChat wrote:

> Knock Knock!
`)
	expected := "Who's There?"
	if result != expected {
		t.Errorf("expected %q got %q", expected, result)
	}
}
