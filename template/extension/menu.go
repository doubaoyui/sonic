package extension

import (
	"context"

	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/model/vo"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/template"
)

type menuExtension struct {
	MenuService   service.MenuService
	OptionService service.OptionService
	Template      *template.Template
}

func RegisterMenuFunc(template *template.Template, menuService service.MenuService, optionService service.OptionService) {
	m := &menuExtension{
		Template:      template,
		MenuService:   menuService,
		OptionService: optionService,
	}
	m.addListMenuFunc()
	m.addListMenuAsTree()
	m.addListTeams()
	m.addListMenuByTeam()
	m.addListMenuAsTreeByTeam()
	m.addGetMenuCount()
}

func (m *menuExtension) addListMenuFunc() {
	listMenu := func() ([]*dto.Menu, error) {
		ctx := context.Background()
		listTeam := m.OptionService.GetOrByDefault(ctx, property.DefaultMenuTeam).(string)
		resolvedTeam := listTeam
		if resolvedTeam == "" {
			if teams, err := m.MenuService.ListTeams(ctx); err == nil && len(teams) > 0 {
				for _, team := range teams {
					if team != "" {
						resolvedTeam = team
						break
					}
				}
			}
		}

		menus, err := m.MenuService.ListByTeam(ctx, resolvedTeam, &param.Sort{
			Fields: []string{"priority,asc"},
		})
		if err != nil {
			return nil, err
		}
		if len(menus) == 0 && listTeam != "" {
			if teams, err := m.MenuService.ListTeams(ctx); err == nil && len(teams) > 0 {
				for _, team := range teams {
					if team == "" || team == listTeam {
						continue
					}
					if fallbackMenus, err := m.MenuService.ListByTeam(ctx, team, &param.Sort{Fields: []string{"priority,asc"}}); err == nil && len(fallbackMenus) > 0 {
						menus = fallbackMenus
						break
					}
				}
			}
		}
		menuDTOs := m.MenuService.ConvertToDTOs(ctx, menus)
		return menuDTOs, nil
	}
	m.Template.AddFunc("listMenu", listMenu)
}

func (m *menuExtension) addListMenuAsTree() {
	listMenuAsTree := func() ([]*vo.Menu, error) {
		ctx := context.Background()
		listTeam := m.OptionService.GetOrByDefault(ctx, property.DefaultMenuTeam).(string)
		resolvedTeam := listTeam
		if resolvedTeam == "" {
			if teams, err := m.MenuService.ListTeams(ctx); err == nil && len(teams) > 0 {
				for _, team := range teams {
					if team != "" {
						resolvedTeam = team
						break
					}
				}
			}
		}
		menus, err := m.MenuService.ListAsTreeByTeam(ctx, resolvedTeam, &param.Sort{Fields: []string{"priority,asc"}})
		if err != nil {
			return nil, err
		}
		if len(menus) == 0 && listTeam != "" {
			if teams, err := m.MenuService.ListTeams(ctx); err == nil && len(teams) > 0 {
				for _, team := range teams {
					if team == "" || team == listTeam {
						continue
					}
					if fallback, err := m.MenuService.ListAsTreeByTeam(ctx, team, &param.Sort{Fields: []string{"priority,asc"}}); err == nil && len(fallback) > 0 {
						return fallback, nil
					}
				}
			}
		}
		return menus, nil
	}
	m.Template.AddFunc("listMenuAsTree", listMenuAsTree)
}

func (m *menuExtension) addListTeams() {
	listMenuTeams := func() ([]string, error) {
		ctx := context.Background()
		teams, err := m.MenuService.ListTeams(ctx)
		return teams, err
	}
	m.Template.AddFunc("listMenuTeams", listMenuTeams)
}

func (m *menuExtension) addListMenuByTeam() {
	listMenuByTeam := func(team string) ([]*dto.Menu, error) {
		ctx := context.Background()
		menus, err := m.MenuService.ListByTeam(ctx, team, &param.Sort{
			Fields: []string{"priority,asc"},
		})
		if err != nil {
			return nil, err
		}
		menuDTOs := m.MenuService.ConvertToDTOs(ctx, menus)
		return menuDTOs, nil
	}
	m.Template.AddFunc("listMenuByTeam", listMenuByTeam)
}

func (m *menuExtension) addListMenuAsTreeByTeam() {
	listMenuAsTreeByTeam := func(team string) ([]*vo.Menu, error) {
		ctx := context.Background()
		menus, err := m.MenuService.ListAsTreeByTeam(ctx, team, &param.Sort{
			Fields: []string{"priority,asc"},
		})
		return menus, err
	}
	m.Template.AddFunc("listMenuAsTreeByTeam", listMenuAsTreeByTeam)
}

func (m *menuExtension) addGetMenuCount() {
	getMenuCount := func() (int64, error) {
		ctx := context.Background()
		return m.MenuService.GetMenuCount(ctx)
	}
	m.Template.AddFunc("getMenuCount", getMenuCount)
}
