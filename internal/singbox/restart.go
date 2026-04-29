package singbox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type RestartOptions struct {
	Enabled       bool
	Driver        string
	DockerSocket  string
	ContainerName string
	Command       string
	Args          []string
	Timeout       time.Duration
}

type RestartResult struct {
	Enabled    bool   `json:"enabled"`
	Executed   bool   `json:"executed"`
	Success    bool   `json:"success"`
	Skipped    bool   `json:"skipped"`
	DurationMS int64  `json:"duration_ms"`
	Message    string `json:"message"`
}

func Restart(ctx context.Context, options RestartOptions) RestartResult {
	result := RestartResult{Enabled: options.Enabled}
	if !options.Enabled {
		result.Skipped = true
		result.Message = "restart disabled"
		return result
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	result.Executed = true
	err := restartWithDriver(ctx, options)
	result.DurationMS = time.Since(started).Milliseconds()
	if err == nil {
		result.Success = true
		result.Message = "restart completed"
		return result
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.Message = "restart timed out"
		return result
	}
	result.Message = "restart failed"
	return result
}

func restartWithDriver(ctx context.Context, options RestartOptions) error {
	switch strings.ToLower(strings.TrimSpace(options.Driver)) {
	case "", "docker":
		return restartDockerContainer(ctx, options)
	case "command":
		return restartCommand(ctx, options)
	default:
		return fmt.Errorf("unsupported restart driver")
	}
}

func restartDockerContainer(ctx context.Context, options RestartOptions) error {
	socketPath := strings.TrimSpace(options.DockerSocket)
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}
	containerName := strings.TrimSpace(options.ContainerName)
	if containerName == "" {
		return fmt.Errorf("sing-box container name is required")
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{Transport: transport}
	endpoint := "http://docker/containers/" + url.PathEscape(containerName) + "/restart?t=10"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("docker restart returned status %d", response.StatusCode)
	}
	return nil
}

func restartCommand(ctx context.Context, options RestartOptions) error {
	command := strings.TrimSpace(options.Command)
	if command == "" {
		return fmt.Errorf("restart command is not configured")
	}
	return exec.CommandContext(ctx, command, options.Args...).Run()
}
