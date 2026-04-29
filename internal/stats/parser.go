package stats

import (
	"sort"
	"strings"

	"github.com/winds18/FluxGate/internal/store"
)

type Counter struct {
	Name  string
	Value int64
}

func SamplesFromV2RayCounters(sampledAtUnix int64, counters []Counter) []store.RecordTrafficSampleInput {
	type aggregate struct {
		metricType string
		metricName string
		upload     int64
		download   int64
	}

	aggregates := map[string]aggregate{}
	for _, counter := range counters {
		if counter.Value < 0 {
			continue
		}
		metricType, metricName, direction, ok := parseV2RayCounterName(counter.Name)
		if !ok {
			continue
		}
		key := metricType + "\x00" + metricName
		item := aggregates[key]
		item.metricType = metricType
		item.metricName = metricName
		switch direction {
		case "uplink":
			item.upload = counter.Value
		case "downlink":
			item.download = counter.Value
		}
		aggregates[key] = item
	}

	keys := make([]string, 0, len(aggregates))
	for key := range aggregates {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	samples := make([]store.RecordTrafficSampleInput, 0, len(keys))
	for _, key := range keys {
		item := aggregates[key]
		samples = append(samples, store.RecordTrafficSampleInput{
			SampledAt:        unixUTC(sampledAtUnix),
			MetricType:       item.metricType,
			MetricName:       item.metricName,
			RawValueUpload:   item.upload,
			RawValueDownload: item.download,
		})
	}
	return samples
}

func parseV2RayCounterName(name string) (string, string, string, bool) {
	parts := strings.Split(strings.TrimSpace(name), ">>>")
	if len(parts) != 4 || parts[2] != "traffic" {
		return "", "", "", false
	}
	metricType := strings.ToLower(strings.TrimSpace(parts[0]))
	metricName := strings.TrimSpace(parts[1])
	direction := strings.ToLower(strings.TrimSpace(parts[3]))
	if metricName == "" {
		return "", "", "", false
	}
	if metricType != "user" && metricType != "inbound" && metricType != "outbound" {
		return "", "", "", false
	}
	if direction != "uplink" && direction != "downlink" {
		return "", "", "", false
	}
	return metricType, metricName, direction, true
}
