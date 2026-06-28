package app

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestPromptTreatsControlCAsInterrupt(t *testing.T) {
	a := &App{
		in:  bufio.NewReader(strings.NewReader("\x03\n")),
		out: io.Discard,
	}

	_, err := a.prompt("Choose: ")
	if !errors.Is(err, errInputInterrupted) {
		t.Fatalf("expected interrupt error, got %v", err)
	}
	if !a.done {
		t.Fatal("expected app to be marked done")
	}
}
