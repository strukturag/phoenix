package phoenix_test

import (
	"os"
	"os/signal"
	"testing"
	"time"

	phoenix "."
)

func withTestServer(t *testing.T, server phoenix.Server, runFunc phoenix.RunFunc, runTest func()) {
	self, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to get process handle for self: %v", err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	go func() {
		if err := server.Run(runFunc); err != nil {
			t.Fatalf("Unexpected error starting server: %v", err)
		} else {
			t.Log("Server shutdown cleanly")
		}
	}()

	// TODO(lcooper): If and when we can signal that we've bound all of our
	// sockets, use that instead of waiting.
	time.Sleep(1000 * time.Millisecond)

	runTest()

	if err := self.Signal(os.Interrupt); err == nil {
		<- sig
	} else {
		t.Fatalf("Failed to kill server, inconsistant state will result: %v", err)
	}
}
