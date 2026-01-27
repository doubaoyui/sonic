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
		return fmt.Sprintf(val, args...)
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
	},
	"en": {
		"landing.download_windows": "Download for Windows",
		"landing.download_macos":   "Download for macOS",
		"landing.download_now":     "Download now",
		"landing.clone_source":     "Clone source",
		"landing.available_for":    "Available for macOS, Linux, and Windows",
	},
}
