package stats

import "testing"

func TestSamplesFromV2RayCounters(t *testing.T) {
	samples := SamplesFromV2RayCounters(1777443000, []Counter{
		{Name: "user>>>fg_u_1_t_1>>>traffic>>>uplink", Value: 120},
		{Name: "user>>>fg_u_1_t_1>>>traffic>>>downlink", Value: 340},
		{Name: "inbound>>>vn-FluxGate-HK>>>traffic>>>uplink", Value: 10},
		{Name: "outbound>>>up_42>>>traffic>>>downlink", Value: 77},
		{Name: "outbound>>>up_42>>>traffic>>>uplink", Value: 55},
		{Name: "system>>>ignored>>>traffic>>>uplink", Value: 1},
		{Name: "user>>>negative>>>traffic>>>uplink", Value: -1},
	})

	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %+v", samples)
	}
	if samples[0].MetricType != "inbound" || samples[0].MetricName != "vn-FluxGate-HK" || samples[0].RawValueUpload != 10 || samples[0].RawValueDownload != 0 {
		t.Fatalf("unexpected inbound sample: %+v", samples[0])
	}
	if samples[1].MetricType != "outbound" || samples[1].MetricName != "up_42" || samples[1].RawValueUpload != 55 || samples[1].RawValueDownload != 77 {
		t.Fatalf("unexpected outbound sample: %+v", samples[1])
	}
	if samples[2].MetricType != "user" || samples[2].MetricName != "fg_u_1_t_1" || samples[2].RawValueUpload != 120 || samples[2].RawValueDownload != 340 {
		t.Fatalf("unexpected user sample: %+v", samples[2])
	}
	if samples[2].SampledAt.IsZero() {
		t.Fatalf("expected sampled_at to be populated")
	}
}
