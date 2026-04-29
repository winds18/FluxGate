package singbox

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestRestartSkipsWhenDisabled(t *testing.T) {
	result := Restart(context.Background(), RestartOptions{Enabled: false})
	if result.Enabled || result.Executed || !result.Skipped || result.Success {
		t.Fatalf("unexpected disabled restart result: %+v", result)
	}
}

func TestRestartRunsConfiguredCommand(t *testing.T) {
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true command is not available")
	}
	result := Restart(context.Background(), RestartOptions{
		Enabled: true,
		Driver:  "command",
		Command: truePath,
		Timeout: time.Second,
	})
	if !result.Enabled || !result.Executed || !result.Success || result.Skipped {
		t.Fatalf("unexpected successful restart result: %+v", result)
	}
}

func TestRestartReportsCommandFailure(t *testing.T) {
	falsePath, err := exec.LookPath("false")
	if err != nil {
		t.Skip("false command is not available")
	}
	result := Restart(context.Background(), RestartOptions{
		Enabled: true,
		Driver:  "command",
		Command: falsePath,
		Timeout: time.Second,
	})
	if !result.Enabled || !result.Executed || result.Success || result.Skipped {
		t.Fatalf("unexpected failed restart result: %+v", result)
	}
}

func TestRestartDockerDriverCallsDockerSocket(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), "fg-restart-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
	defer os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	defer listener.Close()

	requests := make(chan string, 1)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	})}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Close()

	result := Restart(context.Background(), RestartOptions{
		Enabled:       true,
		Driver:        "docker",
		DockerSocket:  socketPath,
		ContainerName: "fluxgate-sing-box",
		Timeout:       time.Second,
	})
	if !result.Enabled || !result.Executed || !result.Success || result.Skipped {
		t.Fatalf("unexpected docker restart result: %+v", result)
	}
	select {
	case got := <-requests:
		if got != "/containers/fluxgate-sing-box/restart?t=10" {
			t.Fatalf("unexpected docker restart request: %s", got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected docker restart request")
	}
}
