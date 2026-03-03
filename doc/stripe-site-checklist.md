# Stripe Review Site Checklist (Sonic)

Goal: make the public website pass Stripe review requirements for a software/digital product.

## Minimum Pages (Public)
- Home: what the product is, key value, who it’s for.
- Download: delivery method (download), platforms, version info.
- Pricing: plans, currency, billing cycle, free trial, cancellation.
- Contact: support email, company/owner name, location (city/country).
- Terms: service terms, subscription terms, cancellation, refund policy.
- Privacy: privacy policy, data collection/processing.

Recommended:
- Refund: if you prefer a dedicated refund policy page.
- FAQ: common questions about billing/refunds/usage.

## Sonic Setup
- Pages: create Sheets with slugs (ROOT permalink recommended):
  - `download`, `pricing`, `docs`, `faq`, `contact`, `about`, `terms`, `privacy`
- Menu: `Appearance -> Menus`, add URLs:
  - `/`, `/download`, `/pricing`, `/docs`, `/faq`, `/contact`, `/about`
- Footer: keep `Privacy` + `Terms` + `Contact` always visible.

## i18n (ZH/EN)
- Menu name: use `中文||English` in the menu `name` field.
- Sheet title: use `中文||English` in `Title`.
- Sheet content: include both languages using markers:
  - `<!--lang:zh-->` ... `<!--lang:en-->` ...

Example:
```md
<!--lang:zh-->
## 定价
这里是中文说明。

<!--lang:en-->
## Pricing
English copy here.
```

## Operational Notes
- Use HTTPS in production (Stripe checks this).
- Ensure links don’t 404.
- Ensure policies are accessible without login.

