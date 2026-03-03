package extension

import (
	"fmt"
	"strings"

	"github.com/go-sonic/sonic/template"
)

// RegisterI18nFunc registers a small translation helper for templates:
//   {{t .lang "landing.download_windows"}}
//   {{t .lang "landing.download_windows" "fallback text"}}
//
// Supported languages out of the box: zh/en.
func RegisterI18nFunc(t *template.Template) {
	t.AddFunc("t", func(lang string, key string, args ...interface{}) string {
		lang = normalizeLang(lang)
		val, ok := i18nTable[lang][key]
		if !ok || val == "" {
			// Fallback: default language (zh), then optional fallback arg, then key.
			if v, ok2 := i18nTable["zh"][key]; ok2 && v != "" {
				val = v
			} else if len(args) > 0 {
				if s, ok3 := args[0].(string); ok3 && s != "" {
					return s
				}
				return fmt.Sprint(args[0])
			} else {
				return key
			}
		}
		if len(args) == 0 {
			return val
		}
		// Optional formatting: {{t .lang "x" .a .b}}
		// If the translation does not use formatting verbs, ignore extra args.
		formatted := fmt.Sprintf(val, args...)
		if strings.Contains(formatted, "%!(EXTRA") || strings.Contains(formatted, "%!(") {
			return val
		}
		return formatted
	})

	// i18nText picks a language variant from a single string without needing schema/UI changes.
	//
	// Convention:
	// - "中文||English" (recommended)
	// - "中文|English"  (fallback)
	//
	// Examples:
	// - Menu name: "下载||Download"
	// - Button label: "立即下载||Download"
	t.AddFunc("i18nText", func(lang string, value string) string {
		value = strings.TrimSpace(value)
		if value == "" {
			return ""
		}
		// Normalize common "bar" variants users may input from different IMEs.
		// This keeps the admin UI simple: user can type "中文||English" with many bar-looking characters.
		var b strings.Builder
		b.Grow(len(value))
		for _, r := range value {
			switch r {
			// Single-bar variants -> |
			case '｜', '¦', '丨', '∣', '│', '┃', '￨', '︱', '︲', 'ǀ', '⎮', '❘', '❙':
				b.WriteRune('|')
			// Double-bar variants -> ||
			case '‖', '∥':
				b.WriteString("||")
			default:
				b.WriteRune(r)
			}
		}
		value = b.String()

		base := normalizeLang(lang)

		sep := "||"
		if !strings.Contains(value, sep) && strings.Contains(value, "|") {
			sep = "|"
		}
		if !strings.Contains(value, sep) {
			return value
		}
		parts := strings.SplitN(value, sep, 2)
		zh := strings.TrimSpace(parts[0])
		en := ""
		if len(parts) > 1 {
			en = strings.TrimSpace(parts[1])
		}
		if base == "en" {
			if en != "" {
				return en
			}
			return zh
		}
		// Default to zh.
		if zh != "" {
			return zh
		}
		return en
	})

	// i18nBlock extracts a language-specific block from a single HTML/Markdown-derived string.
	//
	// Use HTML comments as markers (works well with Markdown editors):
	//   <!--lang:zh-->
	//   ...content...
	//   <!--lang:en-->
	//   ...content...
	//
	// Only the requested language block is returned. If no markers exist, the original value is returned.
	t.AddFunc("i18nBlock", func(lang string, value string) string {
		value = strings.TrimSpace(value)
		if value == "" {
			return ""
		}
		base := normalizeLang(lang)

		// Normalize marker search by lowering a copy.
		lower := strings.ToLower(value)
		findMarker := func(marker string) (start int, end int) {
			// Allow whitespace: <!-- lang:zh -->, <!--lang:zh-->
			idx := strings.Index(lower, marker)
			if idx < 0 {
				return -1, -1
			}
			// Find the end of comment marker.
			gt := strings.Index(lower[idx:], "-->")
			if gt < 0 {
				return -1, -1
			}
			return idx, idx + gt + 3
		}

		zhMarkerA, zhMarkerB := findMarker("<!--lang:zh")
		enMarkerA, enMarkerB := findMarker("<!--lang:en")
		// Also support a common spaced form: "<!-- lang:zh"
		if zhMarkerA < 0 {
			zhMarkerA, zhMarkerB = findMarker("<!-- lang:zh")
		}
		if enMarkerA < 0 {
			enMarkerA, enMarkerB = findMarker("<!-- lang:en")
		}
		if zhMarkerA < 0 && enMarkerA < 0 {
			return value
		}

		// Compute the bounds of blocks by ordering markers.
		type mark struct {
			lang     string
			startIdx int
			endIdx   int
		}
		marks := make([]mark, 0, 2)
		if zhMarkerA >= 0 {
			marks = append(marks, mark{lang: "zh", startIdx: zhMarkerA, endIdx: zhMarkerB})
		}
		if enMarkerA >= 0 {
			marks = append(marks, mark{lang: "en", startIdx: enMarkerA, endIdx: enMarkerB})
		}
		if len(marks) == 1 {
			only := marks[0]
			block := strings.TrimSpace(value[only.endIdx:])
			if base == only.lang {
				return block
			}
			return ""
		}
		// Sort the two markers.
		first, second := marks[0], marks[1]
		if second.startIdx < first.startIdx {
			first, second = second, first
		}

		firstBlock := strings.TrimSpace(value[first.endIdx:second.startIdx])
		secondBlock := strings.TrimSpace(value[second.endIdx:])
		if base == first.lang {
			return firstBlock
		}
		if base == second.lang {
			return secondBlock
		}
		// Unknown lang: fallback to zh if present, else first.
		if first.lang == "zh" {
			return firstBlock
		}
		if second.lang == "zh" {
			return secondBlock
		}
		return firstBlock
	})

	// optStr reads a string-ish value from a map (e.g. theme settings) safely.
	// Missing / nil returns "".
	t.AddFunc("optStr", func(m map[string]interface{}, key string) string {
		if m == nil || key == "" {
			return ""
		}
		v, ok := m[key]
		if !ok || v == nil {
			return ""
		}
		switch vv := v.(type) {
		case string:
			return vv
		case fmt.Stringer:
			return vv.String()
		default:
			return fmt.Sprint(v)
		}
	})
}

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return "zh"
	}
	lang = strings.ToLower(lang)
	if idx := strings.IndexAny(lang, "-_"); idx >= 0 {
		lang = lang[:idx]
	}
	switch lang {
	case "zh":
		return "zh"
	case "en":
		return "en"
	default:
		return lang
	}
}

var i18nTable = map[string]map[string]string{
	"zh": {
		"landing.download_windows": "下载 Windows",
		"landing.download_macos":   "下载 macOS",
		"landing.download_now":     "立即下载",
		"landing.clone_source":     "克隆源码",
		"landing.available_for":    "支持 macOS、Linux、Windows",
		"footer.legal":             "条款",
		"footer.support":           "支持",
		"footer.privacy":           "隐私政策",
		"footer.terms":             "使用条款",
		"footer.contact":           "联系我们",
		"footer.refund":            "退款政策",
		"pricing.one_time":         "一次性",
		"pricing.credits":          "积分",
		"pricing.get_started":      "开始使用",
		"pricing.buy_now":          "立即购买",
	},
	"en": {
		"landing.download_windows": "Download for Windows",
		"landing.download_macos":   "Download for macOS",
		"landing.download_now":     "Download now",
		"landing.clone_source":     "Clone source",
		"landing.available_for":    "Available for macOS, Linux, and Windows",
		"footer.legal":             "Legal",
		"footer.support":           "Support",
		"footer.privacy":           "Privacy Policy",
		"footer.terms":             "Terms of Use",
		"footer.contact":           "Contact",
		"footer.refund":            "Refund Policy",
		"pricing.one_time":         "One‑time",
		"pricing.credits":          "CREDITS",
		"pricing.get_started":      "Get Started",
		"pricing.buy_now":          "Buy Now",
	},
}
