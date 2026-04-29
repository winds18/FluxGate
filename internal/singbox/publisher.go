package singbox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/winds18/FluxGate/internal/store"
)

type ConfigStore interface {
	ListTokens(ctx context.Context) ([]store.TokenWithAccount, error)
	ListVirtualNodes(ctx context.Context) ([]store.VirtualNode, error)
	ListNodes(ctx context.Context) ([]store.Node, error)
	ListPolicies(ctx context.Context) ([]store.Policy, error)
}

type Publisher struct {
	Store              ConfigStore
	ConfigPath         string
	PreviousConfigPath string
	AutoRestart        bool
	RestartOptions     RestartOptions
	Logger             *slog.Logger
}

func BuildConfigFromStore(ctx context.Context, store ConfigStore) (Config, error) {
	tokens, err := store.ListTokens(ctx)
	if err != nil {
		return Config{}, err
	}
	virtualNodes, err := store.ListVirtualNodes(ctx)
	if err != nil {
		return Config{}, err
	}
	upstreamNodes, err := store.ListNodes(ctx)
	if err != nil {
		return Config{}, err
	}
	policies, err := store.ListPolicies(ctx)
	if err != nil {
		return Config{}, err
	}
	return BuildConfigWithPolicies(tokens, virtualNodes, upstreamNodes, policies), nil
}

func (p Publisher) TriggerConfigPublish(ctx context.Context, reason string) error {
	if p.Store == nil {
		return fmt.Errorf("sing-box config store is required")
	}
	config, err := BuildConfigFromStore(ctx, p.Store)
	if err != nil {
		return err
	}
	result, err := PublishConfig(config, p.ConfigPath, p.PreviousConfigPath)
	if err != nil {
		return err
	}
	if p.AutoRestart {
		options := p.RestartOptions
		options.Enabled = true
		restart := Restart(ctx, options)
		result.Restart = &restart
		result.RestartRequired = !restart.Success
		if !restart.Success {
			p.logPublishResult(reason, result)
			return fmt.Errorf("sing-box restart failed: %s", restart.Message)
		}
	}
	p.logPublishResult(reason, result)
	return nil
}

func (p Publisher) logPublishResult(reason string, result PublishResult) {
	if p.Logger == nil {
		return
	}
	p.Logger.Info("sing-box config publish triggered",
		"reason", reason,
		"published", result.Published,
		"config_hash", result.ConfigHash,
		"previous_saved", result.PreviousSaved,
		"restart_required", result.RestartRequired,
		"restart_attached", result.Restart != nil,
	)
}
