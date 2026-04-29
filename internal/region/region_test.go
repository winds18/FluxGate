package region

import "testing"

func TestNormalizeChinaRegions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		hints   []string
		want    string
	}{
		{name: "hong kong chinese", hints: []string{"🇭🇰3香港集群-全网优化(AnyTLS)"}, want: ChinaHongKong},
		{name: "hong kong alias", current: "HK", want: ChinaHongKong},
		{name: "taiwan simplified", hints: []string{"5台湾-联通/移动(AnyTLS)"}, want: ChinaTaiwan},
		{name: "taiwan traditional", hints: []string{"台灣 01"}, want: ChinaTaiwan},
		{name: "taipei english", hints: []string{"Taipei relay"}, want: ChinaTaiwan},
		{name: "route destination wins", hints: []string{"香港-美国-专线"}, want: "美国"},
		{name: "route destination ignores product suffix", hints: []string{"香港-专线"}, want: ChinaHongKong},
		{name: "route destination can be china region", hints: []string{"美国→台湾"}, want: ChinaTaiwan},
		{name: "known country", hints: []string{"🇯🇵7日本-专线(AnyTLS)"}, want: "日本"},
		{name: "known city", hints: []string{"旧金山 01"}, want: "美国"},
		{name: "preserve unknown", current: "东南亚", hints: []string{"unknown"}, want: "东南亚"},
		{name: "empty unknown", hints: []string{"unknown"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.current, tt.hints...); got != tt.want {
				t.Fatalf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}
