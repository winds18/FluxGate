package stats

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

type Collector interface {
	CollectCounters(ctx context.Context) ([]Counter, error)
}

type TrafficRecorder interface {
	RecordTrafficSamples(ctx context.Context, inputs []store.RecordTrafficSampleInput) ([]store.TrafficSample, error)
}

type Poller struct {
	Collector Collector
	Recorder  TrafficRecorder
	Logger    *slog.Logger
	Now       func() time.Time
}

type PollResult struct {
	SampledAt     time.Time
	CounterCount  int
	SampleCount   int
	RecordedCount int
}

func (p Poller) PollOnce(ctx context.Context) (PollResult, error) {
	if p.Collector == nil {
		return PollResult{}, errors.New("stats collector is required")
	}
	if p.Recorder == nil {
		return PollResult{}, errors.New("traffic recorder is required")
	}

	sampledAt := p.now()
	counters, err := p.Collector.CollectCounters(ctx)
	if err != nil {
		return PollResult{SampledAt: sampledAt}, err
	}
	samples := SamplesFromV2RayCounters(sampledAt.Unix(), counters)
	result := PollResult{
		SampledAt:    sampledAt,
		CounterCount: len(counters),
		SampleCount:  len(samples),
	}
	if len(samples) == 0 {
		return result, nil
	}
	recorded, err := p.Recorder.RecordTrafficSamples(ctx, samples)
	if err != nil {
		return result, err
	}
	result.RecordedCount = len(recorded)
	return result, nil
}

func (p Poller) RunScheduler(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	p.pollAndLog(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollAndLog(ctx)
		}
	}
}

func (p Poller) pollAndLog(ctx context.Context) {
	result, err := p.PollOnce(ctx)
	if err != nil {
		if p.Logger != nil {
			p.Logger.Warn("stats poll failed", "error", err)
		}
		return
	}
	if p.Logger != nil {
		p.Logger.Info("stats poll completed",
			"counters", result.CounterCount,
			"samples", result.SampleCount,
			"recorded", result.RecordedCount,
			"sampled_at", result.SampledAt.Format(time.RFC3339),
		)
	}
}

func (p Poller) now() time.Time {
	if p.Now != nil {
		return p.Now().UTC()
	}
	return time.Now().UTC()
}
