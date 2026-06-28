package model

import "testing"

func TestSentenceFormatterApply(t *testing.T) {
	f := SentenceFormatter{Color: "#00557f"}
	got := f.Apply("The system will collapse under pressure.", "collapse")
	want := "The system will <span style=\"color:#00557f;\"><b>collapse</b></span> under pressure."

	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
