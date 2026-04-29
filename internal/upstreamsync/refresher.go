package upstreamsync

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/store"
	"github.com/winds18/FluxGate/internal/upstream"
)

type FetchFunc func(context.Context, string) (string, error)

type Refresher struct {
	Store *store.Store
	Fetch FetchFunc
}

type Summary struct {
	Checked     int `json:"checked"`
	Succeeded   int `json:"succeeded"`
	Failed      int `json:"failed"`
	Imported    int `json:"imported"`
	Updated     int `json:"updated"`
	Skipped     int `json:"skipped"`
	Inactivated int `json:"inactivated"`
}

func (r Refresher) RefreshSource(ctx context.Context, source store.Source) (store.ImportResult, error) {
	content := source.RawContent
	if strings.TrimSpace(source.URL) != "" {
		fetched, err := r.fetch(ctx, source.URL)
		if err != nil {
			return store.ImportResult{}, err
		}
		content = fetched
	}
	normalized, err := upstream.NormalizeContent(content)
	if err != nil {
		return store.ImportResult{}, err
	}
	if err := r.Store.UpdateSourceRawContent(ctx, source.ID, normalized); err != nil {
		return store.ImportResult{}, err
	}
	result, err := r.Store.ImportNodes(ctx, store.ImportNodesInput{
		SourceID:            source.ID,
		Content:             normalized,
		MarkMissingInactive: true,
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r Refresher) SyncDue(ctx context.Context, now time.Time, limit int, logger *slog.Logger) (Summary, error) {
	sources, err := r.Store.ListDueSubscriptionSources(ctx, now, limit)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{Checked: len(sources)}
	for _, source := range sources {
		result, err := r.RefreshSource(ctx, source)
		if err != nil {
			summary.Failed++
			_ = r.Store.SetSourceSyncError(ctx, source.ID, err.Error())
			if logger != nil {
				logger.Info("scheduled source refresh failed",
					"source_id", source.ID,
					"source_type", source.Type,
					"reason", err.Error(),
				)
			}
			continue
		}
		summary.Succeeded++
		summary.Imported += result.Imported
		summary.Updated += result.Updated
		summary.Skipped += result.Skipped
		summary.Inactivated += result.Inactivated
		if logger != nil {
			logger.Info("scheduled source refresh succeeded",
				"source_id", source.ID,
				"source_type", source.Type,
				"imported", result.Imported,
				"updated", result.Updated,
				"skipped", result.Skipped,
				"inactivated", result.Inactivated,
			)
		}
	}
	return summary, nil
}

func (r Refresher) RunScheduler(ctx context.Context, interval time.Duration, limit int, logger *slog.Logger) {
	if interval <= 0 {
		return
	}
	run := func() {
		summary, err := r.SyncDue(ctx, time.Now().UTC(), limit, logger)
		if err != nil {
			if logger != nil {
				logger.Info("scheduled source sync failed", "reason", err.Error())
			}
			return
		}
		if logger != nil && summary.Checked > 0 {
			logger.Info("scheduled source sync completed",
				"checked", summary.Checked,
				"succeeded", summary.Succeeded,
				"failed", summary.Failed,
				"imported", summary.Imported,
				"updated", summary.Updated,
				"skipped", summary.Skipped,
				"inactivated", summary.Inactivated,
			)
		}
	}

	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func (r Refresher) fetch(ctx context.Context, sourceURL string) (string, error) {
	if r.Fetch != nil {
		return r.Fetch(ctx, sourceURL)
	}
	return upstream.Fetcher{}.Fetch(ctx, sourceURL)
}
