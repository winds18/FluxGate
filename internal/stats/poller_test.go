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
	inputs []store.RecordTrafficSampleInput
	err    error
}

func (f *fakeRecorder) RecordTrafficSamples(_ context.Context, inputs []store.RecordTrafficSampleInput) ([]store.TrafficSample, error) {
	f.inputs = append(f.inputs, inputs...)
	if f.err != nil {
		return nil, f.err
	}
	samples := make([]store.TrafficSample, len(inputs))
	return samples, nil
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
