package region

import (
	"strings"
	"unicode"
)

const (
	ChinaHongKong = "🇨🇳中国|香港"
	ChinaTaiwan   = "🇨🇳中国|台湾"
)

type alias struct {
	name     string
	needles  []string
	keywords []string
	tokens   []string
}

var aliases = []alias{
	{name: ChinaHongKong, needles: []string{"🇭🇰", "香港", "hong kong", "hongkong"}, tokens: []string{"hk", "hkg"}},
	{name: ChinaTaiwan, needles: []string{"🇹🇼", "台湾", "台灣", "台北", "taiwan", "taipei"}, tokens: []string{"tw", "tpe"}},
	{name: "美国", needles: []string{"🇺🇸", "美国", "美西", "美东", "旧金山", "华盛顿", "维加斯", "洛杉矶", "西雅图", "纽约", "united states", "unitedstates", "america", "san francisco", "sanfrancisco", "washington", "las vegas", "lasvegas", "los angeles", "losangeles", "new york", "newyork", "seattle"}, tokens: []string{"us", "usa", "sfo", "sjc", "lax", "iad", "nyc", "sea"}},
	{name: "加拿大", needles: []string{"🇨🇦", "加拿大", "canada"}, tokens: []string{"ca", "yvr", "yyz"}},
	{name: "土耳其", needles: []string{"🇹🇷", "土耳其", "turkey"}},
	{name: "俄罗斯", needles: []string{"🇷🇺", "俄罗斯", "russia"}},
	{name: "越南", needles: []string{"🇻🇳", "越南", "vietnam"}},
	{name: "印尼", needles: []string{"🇮🇩", "印尼", "印度尼西亚", "indonesia"}},
	{name: "日本", needles: []string{"🇯🇵", "日本", "东京", "大阪", "japan", "tokyo", "osaka"}, tokens: []string{"jp", "jpn", "nrt", "hnd", "kix"}},
	{name: "韩国", needles: []string{"🇰🇷", "韩国", "韓國", "首尔", "首爾", "korea", "seoul"}, tokens: []string{"kr", "kor", "icn"}},
	{name: "新加坡", needles: []string{"🇸🇬", "新加坡", "singapore"}, tokens: []string{"sg", "sin"}},
	{name: "澳洲", needles: []string{"🇦🇺", "澳洲", "澳大利亚", "澳大利亞", "australia", "sydney"}, tokens: []string{"au", "syd"}},
	{name: "阿联酋", needles: []string{"🇦🇪", "迪拜", "阿联酋", "阿聯酋", "dubai", "uae"}},
	{name: "印度", needles: []string{"🇮🇳", "印度", "india"}},
	{name: "德国", needles: []string{"🇩🇪", "德国", "德國", "germany"}},
	{name: "英国", needles: []string{"🇬🇧", "英国", "英國", "united kingdom", "unitedkingdom", "britain", "london"}, tokens: []string{"uk", "gb", "lon"}},
	{name: "巴西", needles: []string{"🇧🇷", "巴西", "brazil"}},
	{name: "智利", needles: []string{"🇨🇱", "智利", "chile"}},
	{name: "法国", needles: []string{"🇫🇷", "法国", "法國", "france"}},
	{name: "墨西哥", needles: []string{"🇲🇽", "墨西哥", "mexico"}},
	{name: "荷兰", needles: []string{"🇳🇱", "荷兰", "荷蘭", "netherlands", "holland"}},
	{name: "以色列", needles: []string{"🇮🇱", "以色列", "israel"}},
	{name: "西班牙", needles: []string{"🇪🇸", "西班牙", "spain"}},
	{name: "阿根廷", needles: []string{"🇦🇷", "阿根廷", "argentina"}},
	{name: "乌克兰", needles: []string{"🇺🇦", "乌克兰", "烏克蘭", "ukraine"}},
	{name: "瑞士", needles: []string{"🇨🇭", "瑞士", "switzerland"}},
	{name: "南非", needles: []string{"🇿🇦", "南非", "约翰内斯堡", "約翰內斯堡", "south africa", "southafrica", "johannesburg"}},
	{name: "马来西亚", needles: []string{"🇲🇾", "马来西亚", "馬來西亞", "malaysia"}},
	{name: "菲律宾", needles: []string{"🇵🇭", "菲律宾", "菲律賓", "philippines"}},
	{name: "泰国", needles: []string{"🇹🇭", "泰国", "泰國", "thailand"}},
}

func Normalize(current string, hints ...string) string {
	current = strings.TrimSpace(current)
	for _, hint := range hints {
		if matched := matchRouteDestination(hint); matched != "" {
			return matched
		}
	}
	if matched := match(current); matched != "" {
		return matched
	}
	for _, hint := range hints {
		if matched := match(hint); matched != "" {
			return matched
		}
	}
	return current
}

func match(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, item := range aliases {
		if item.matches(value) {
			return item.name
		}
	}
	return ""
}

func matchRouteDestination(value string) string {
	parts := splitRouteParts(value)
	if len(parts) < 2 {
		return ""
	}
	for i := len(parts) - 1; i > 0; i-- {
		if matched := match(parts[i]); matched != "" {
			return matched
		}
	}
	return ""
}

func splitRouteParts(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case '-', '–', '—', '→', '➡', '➜', '>', '至', '到':
			return true
		default:
			return false
		}
	})
}

func (a alias) matches(value string) bool {
	lower := strings.ToLower(value)
	compact := strings.ReplaceAll(lower, " ", "")
	for _, needle := range a.needles {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	for _, keyword := range a.keywords {
		if strings.Contains(compact, strings.ToLower(strings.ReplaceAll(keyword, " ", ""))) {
			return true
		}
	}
	for _, token := range a.tokens {
		if hasToken(lower, token) {
			return true
		}
	}
	return false
}

func hasToken(value, token string) bool {
	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	}) {
		if part == token {
			return true
		}
	}
	return false
}
