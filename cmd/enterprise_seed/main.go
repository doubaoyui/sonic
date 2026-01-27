package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/config"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/dal"
	"github.com/go-sonic/sonic/event"
	"github.com/go-sonic/sonic/injection"
	sonicLog "github.com/go-sonic/sonic/log"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	_ "github.com/go-sonic/sonic/service/impl"
)

func main() {
	options := injection.GetOptions()
	options = append(options,
		fx.NopLogger,
		fx.Provide(
			sonicLog.NewLogger,
			sonicLog.NewGormLogger,
			event.NewSyncEventBus,
			dal.NewGormDB,
			cache.NewCache,
			config.NewConfig,
		),
		fx.Invoke(func(
			logger *zap.Logger,
			_ *gorm.DB,
			optionService service.OptionService,
			menuService service.MenuService,
			sheetService service.SheetService,
		) error {
			return seedEnterpriseSite(logger, optionService, menuService, sheetService)
		}),
	)

	app := fx.New(options...)
	startCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		panic(err)
	}

	stopCtx, cancelStop := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStop()
	_ = app.Stop(stopCtx)
}

func seedEnterpriseSite(
	logger *zap.Logger,
	optionService service.OptionService,
	menuService service.MenuService,
	sheetService service.SheetService,
) error {
	ctx := context.Background()

	// 1) URL rules for enterprise pages
	if err := optionService.Save(ctx, map[string]string{
		property.DefaultMenuTeam.KeyValue:  "main",
		property.SheetPermalinkType.KeyValue: string(consts.SheetPermaLinkTypeRoot),
		property.PathSuffix.KeyValue:       "",
	}); err != nil {
		return fmt.Errorf("save options: %w", err)
	}

	// 2) Seed sheets (create if missing; otherwise update)
	sheets := []struct {
		Slug  string
		Title string
	}{
		{Slug: "download", Title: "Download"},
		{Slug: "pricing", Title: "Pricing"},
		{Slug: "docs", Title: "Docs"},
		{Slug: "faq", Title: "FAQ"},
		{Slug: "contact", Title: "Contact"},
		{Slug: "about", Title: "About"},
	}

	postDAL := dal.GetQueryByCtx(ctx).Post
	for _, s := range sheets {
		existing, err := postDAL.WithContext(ctx).
			Where(postDAL.Type.Eq(consts.PostTypeSheet), postDAL.Slug.Eq(s.Slug)).
			Take()
		notFound := err != nil && errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !notFound {
			return fmt.Errorf("lookup sheet slug=%s: %w", s.Slug, err)
		}

		original := fmt.Sprintf("# %s\n\n（这里是 %s 页面内容，后续可以在后台编辑替换）", s.Title, s.Title)
		content := fmt.Sprintf("<h1>%s</h1><p>（这里是 %s 页面内容，后续可以在后台编辑替换）</p>", s.Title, s.Title)
		sheetParam := &param.Sheet{
			Title:           s.Title,
			Status:          consts.PostStatusPublished,
			Slug:            s.Slug,
			OriginalContent: original,
			Content:         content,
		}

		if existing == nil || notFound {
			if _, err := sheetService.Create(ctx, sheetParam); err != nil {
				return fmt.Errorf("create sheet slug=%s: %w", s.Slug, err)
			}
			logger.Info("created sheet", zap.String("slug", s.Slug))
			continue
		}
		if _, err := sheetService.Update(ctx, existing.ID, sheetParam); err != nil {
			return fmt.Errorf("update sheet slug=%s: %w", s.Slug, err)
		}
		logger.Info("updated sheet", zap.String("slug", s.Slug))
	}

	// 3) Seed menus for team "main"
	desiredMenus := []param.Menu{
		{Name: "Home", URL: "/", Priority: 1, Target: "_self", Team: "main"},
		{Name: "Download", URL: "/download", Priority: 2, Target: "_self", Team: "main"},
		{Name: "Pricing", URL: "/pricing", Priority: 3, Target: "_self", Team: "main"},
		{Name: "Docs", URL: "/docs", Priority: 4, Target: "_self", Team: "main"},
		{Name: "FAQ", URL: "/faq", Priority: 5, Target: "_self", Team: "main"},
		{Name: "Contact", URL: "/contact", Priority: 6, Target: "_self", Team: "main"},
		{Name: "About", URL: "/about", Priority: 7, Target: "_self", Team: "main"},
	}

	existingMenus, err := menuService.ListByTeam(ctx, "main", &param.Sort{Fields: []string{"priority,asc"}})
	if err != nil {
		return fmt.Errorf("list existing menus: %w", err)
	}
	urlToMenu := make(map[string]int32, len(existingMenus))
	for _, m := range existingMenus {
		urlToMenu[m.URL] = m.ID
	}
	for _, m := range desiredMenus {
		m.ParentID = 0
		if m.Target == "" {
			m.Target = "_self"
		}
		m.Icon = ""
		if id, ok := urlToMenu[m.URL]; ok {
			if _, err := menuService.Update(ctx, id, &m); err != nil {
				return fmt.Errorf("update menu url=%s: %w", m.URL, err)
			}
			logger.Info("updated menu", zap.String("url", m.URL))
			continue
		}
		if _, err := menuService.Create(ctx, &m); err != nil {
			return fmt.Errorf("create menu url=%s: %w", m.URL, err)
		}
		logger.Info("created menu", zap.String("url", m.URL))
	}

	logger.Info("enterprise seed done",
		zap.String("default_menu_team", "main"),
		zap.String("sheet_permalink_type", string(consts.SheetPermaLinkTypeRoot)),
	)
	return nil
}
