package cli

import "testing"

func TestShouldExitCheckOnly(t *testing.T) {
	tests := []struct {
		name        string
		checkOnly   bool
		yes         bool
		interactive bool
		want        bool
	}{
		{name: "script check", checkOnly: true, yes: false, interactive: false, want: true},
		{name: "tty check", checkOnly: true, yes: false, interactive: true, want: false},
		{name: "check yes", checkOnly: true, yes: true, interactive: false, want: false},
		{name: "plain update", checkOnly: false, yes: false, interactive: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldExitCheckOnly(tt.checkOnly, tt.yes, tt.interactive); got != tt.want {
				t.Fatalf("shouldExitCheckOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirmUpdateMessage(t *testing.T) {
	want := "A new update is available (v1.4.0). Would you like to install it?"
	if got := confirmUpdatePrompt("v1.4.0"); got != want {
		t.Fatalf("prompt = %q, want %q", got, want)
	}
}
