package stats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/winds18/FluxGate/internal/store"
)

type fakeCollector struct {
	counters []Counter
	err      error
}

func (f fakeCollector) CollectCounters(context.Context) ([]Counter, error) {
	return f.counters, f.err
}

type fakeRecorder struct {
	inputs                []store.RecordTrafficSampleInput
	err                   error
	configPublishRequired bool
}

func (f *fakeRecorder) RecordTrafficSamples(_ context.Context, inputs []store.RecordTrafficSampleInput) ([]store.TrafficSample, error) {
	f.inputs = append(f.inputs, inputs...)
	if f.err != nil {
		return nil, f.err
	}
	samples := make([]store.TrafficSample, len(inputs))
	return samples, nil
}

func (f *fakeRecorder) RecordTrafficSamplesWithEffects(ctx context.Context, inputs []store.RecordTrafficSampleInput) (store.TrafficRecordResult, error) {
	samples, err := f.RecordTrafficSamples(ctx, inputs)
	if err != nil {
		return store.TrafficRecordResult{}, err
	}
	return store.TrafficRecordResult{
		Samples:               samples,
		ConfigPublishRequired: f.configPublishRequired,
	}, nil
}

type fakeConfigPublishTrigger struct {
	calls  int
	reason string
	err    error
}

func (f *fakeConfigPublishTrigger) TriggerConfigPublish(_ context.Context, reason string) error {
	f.calls++
	f.reason = reason
	return f.err
}

func TestPollOnceRecordsCollectedCounters(t *testing.T) {
	sampledAt := time.Date(2026, 4, 29, 14, 47, 0, 0, time.UTC)
	recorder := &fakeRecorder{}
	poller := Poller{
		Collector: fakeCollector{counters: []Counter{
			{Name: "user>>>fg_u_1_t_1>>>traffic>>>uplink", Value: 120},
			{Name: "user>>>fg_u_1_t_1>>>traffic>>>downlink", Value: 340},
			{Name: "system>>>ignored>>>traffic>>>uplink", Value: 10},
		}},
		Recorder: recorder,
		Now:      func() time.Time { return sampledAt },
	}

	result, err := poller.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("poll once: %v", err)
	}
	if result.CounterCount != 3 || result.SampleCount != 1 || result.RecordedCount != 1 {
		t.Fatalf("unexpected poll result: %+v", result)
	}
	if len(recorder.inputs) != 1 {
		t.Fatalf("expected one recorded sample, got %+v", recorder.inputs)
	}
	input := recorder.inputs[0]
	if input.MetricType != "user" || input.MetricName != "fg_u_1_t_1" {
		t.Fatalf("unexpected recorded metric: %+v", input)
	}
	if input.RawValueUpload != 120 || input.RawValueDownload != 340 {
		t.Fatalf("unexpected raw traffic values: %+v", input)
	}
	if !input.SampledAt.Equal(sampledAt) {
		t.Fatalf("unexpected sampled_at: %s", input.SampledAt)
	}
}

func TestPollOnceSkipsRecorderWhenNoValidSamples(t *testing.T) {
	recorder := &fakeRecorder{}
	poller := Poller{
		Collector: fakeCollector{counters: []Counter{
			{Name: "system>>>ignored>>>traffic>>>uplink", Value: 10},
			{Name: "user>>>negative>>>traffic>>>uplink", Value: -1},
		}},
		Recorder: recorder,
		Now:      func() time.Time { return time.Date(2026, 4, 29, 14, 48, 0, 0, time.UTC) },
	}

	result, err := poller.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("poll once: %v", err)
	}
	if result.CounterCount != 2 || result.SampleCount != 0 || result.RecordedCount != 0 {
		t.Fatalf("unexpected empty poll result: %+v", result)
	}
	if len(recorder.inputs) != 0 {
		t.Fatalf("recorder should not be called for empty samples: %+v", recorder.inputs)
	}
}

func TestPollOnceReturnsCollectorError(t *testing.T) {
	wantErr := errors.New("collect failed")
	poller := Poller{
		Collector: fakeCollector{err: wantErr},
		Recorder:  &fakeRecorder{},
		Now:       func() time.Time { return time.Date(2026, 4, 29, 14, 49, 0, 0, time.UTC) },
	}

	_, err := poller.PollOnce(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected collector error, got %v", err)
	}
}

func TestPollOnceReturnsRecorderError(t *testing.T) {
	wantErr := errors.New("record failed")
	poller := Poller{
		Collector: fakeCollector{counters: []Counter{
			{Name: "outbound>>>up_1>>>traffic>>>uplink", Value: 10},
		}},
		Recorder: &fakeRecorder{err: wantErr},
		Now:      func() time.Time { return time.Date(2026, 4, 29, 14, 50, 0, 0, time.UTC) },
	}

	_, err := poller.PollOnce(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected recorder error, got %v", err)
	}
}

func TestPollOnceTriggersConfigPublishWhenQuotaStateChanges(t *testing.T) {
	recorder := &fakeRecorder{configPublishRequired: true}
	publisher := &fakeConfigPublishTrigger{}
	poller := Poller{
		Collector: fakeCollector{counters: []Counter{
			{Name: "user>>>fg_u_1_t_1>>>traffic>>>uplink", Value: 120},
		}},
		Recorder:             recorder,
		ConfigPublishTrigger: publisher,
		Now:                  func() time.Time { return time.Date(2026, 4, 29, 14, 51, 0, 0, time.UTC) },
	}

	result, err := poller.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("poll once: %v", err)
	}
	if !result.ConfigPublishRequired || !result.ConfigPublishTriggered {
		t.Fatalf("expected config publish trigger in result: %+v", result)
	}
	if publisher.calls != 1 || publisher.reason != "traffic_quota_config_required" {
		t.Fatalf("unexpected publisher calls: calls=%d reason=%q", publisher.calls, publisher.reason)
	}
}

func TestPollOnceReturnsConfigPublishError(t *testing.T) {
	wantErr := errors.New("publish failed")
	poller := Poller{
		Collector: fakeCollector{counters: []Counter{
			{Name: "user>>>fg_u_1_t_1>>>traffic>>>uplink", Value: 120},
		}},
		Recorder:             &fakeRecorder{configPublishRequired: true},
		ConfigPublishTrigger: &fakeConfigPublishTrigger{err: wantErr},
		Now:                  func() time.Time { return time.Date(2026, 4, 29, 14, 52, 0, 0, time.UTC) },
	}

	result, err := poller.PollOnce(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected config publish error, got %v", err)
	}
	if !result.ConfigPublishRequired || result.ConfigPublishTriggered {
		t.Fatalf("unexpected publish error result: %+v", result)
	}
}
