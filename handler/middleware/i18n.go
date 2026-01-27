package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
)

// I18nMiddleware provides a lightweight locale negotiation for templates/pages.
//
// Priority:
// 1) ?lang=xx
// 2) Cookie "sonic_lang"
// 3) Accept-Language header
// 4) blog_locale option (default: "zh")
//
// It sets:
// - gin context key "lang" (normalized, e.g. "zh"/"en")
// - response header "Content-Language"
// - cookie "sonic_lang" when query param is present
type I18nMiddleware struct {
	OptionService service.ClientOptionService
}

func NewI18nMiddleware(optionService service.ClientOptionService) *I18nMiddleware {
	return &I18nMiddleware{
		OptionService: optionService,
	}
}

func (m *I18nMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := normalizeLang(m.getLang(ctx))
		ctx.Set("lang", lang)
		ctx.Header("Content-Language", lang)
		ctx.Next()
	}
}

func (m *I18nMiddleware) getLang(ctx *gin.Context) string {
	if v := strings.TrimSpace(ctx.Query("lang")); v != "" {
		// Persist the explicit lang choice.
		ctx.SetCookie("sonic_lang", v, 60*60*24*365, "/", "", false, false)
		return v
	}
	if v, err := ctx.Cookie("sonic_lang"); err == nil && strings.TrimSpace(v) != "" {
		return v
	}
	if v := strings.TrimSpace(ctx.GetHeader("Accept-Language")); v != "" {
		return v
	}
	// Default from site option.
	if m.OptionService != nil {
		if v := m.OptionService.GetOrByDefault(ctx, property.BlogLocale); v != nil {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return "zh"
}

var acceptLangRe = regexp.MustCompile(`(?i)^\s*([a-z]{2,3})(?:[-_][a-z0-9]{2,8})?`)

func normalizeLang(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "zh"
	}
	// Accept-Language: "en-US,en;q=0.9,zh-CN;q=0.8"
	if idx := strings.Index(raw, ","); idx >= 0 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, ";"); idx >= 0 {
		raw = raw[:idx]
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "zh"
	}
	m := acceptLangRe.FindStringSubmatch(raw)
	if len(m) < 2 {
		return "zh"
	}
	base := strings.ToLower(m[1])
	switch base {
	case "zh":
		return "zh"
	case "en":
		return "en"
	default:
		// Keep base language if it looks reasonable; otherwise fallback.
		if len(base) >= 2 && len(base) <= 3 {
			return base
		}
		return "zh"
	}
}

