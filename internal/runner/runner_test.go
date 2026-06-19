package runner

import "testing"

func TestShutdownRequestForceful(t *testing.T) {
	req := &ShutdownRequest{Forceful: true}
	if !req.Forceful {
		t.Fatal("expected forceful shutdown")
	}
}

func TestShutdownRequestGraceful(t *testing.T) {
	req := &ShutdownRequest{}
	if req.Forceful {
		t.Fatal("expected graceful shutdown by default")
	}
}
