package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "github.com/huzaifanur/ghpm/internal/config"
    "github.com/huzaifanur/ghpm/internal/git"
    "github.com/huzaifanur/ghpm/internal/profile"
    "github.com/huzaifanur/ghpm/pkg/logger"
    "github.com/huzaifanur/ghpm/pkg/version"
)

type UI struct {
    app        fyne.App
    window     fyne.Window
    config     *config.Config
    gitManager *git.Manager
    logger     *logger.Logger

    // UI components
    profileList   *ProfileList
    statusDisplay *StatusDisplay
    toolbar       *Toolbar
    footer        *Footer
    selectedItem  int
    profiles      []*profile.Profile
}

func NewUI(app fyne.App, cfg *config.Config) *UI {
	ui := &UI{
		app:          app,
		config:       cfg,
		gitManager:   git.NewManager(),
		logger:       logger.New(),
		selectedItem: -1,
	}

	ui.setupWindow()
	ui.createComponents()
	ui.buildLayout()
	ui.refresh()

	return ui
}

func (ui *UI) Show() {
	ui.window.ShowAndRun()
}

func (ui *UI) setupWindow() {
	ui.window = ui.app.NewWindow("GitHub Profile Manager")
	ui.window.Resize(fyne.NewSize(900, 700))
	ui.window.CenterOnScreen()
}

func (ui *UI) createComponents() {
    ui.statusDisplay = NewStatusDisplay()
    ui.profileList = NewProfileList(ui)
    ui.toolbar = NewToolbar(ui)
    ui.footer = NewFooter("Made with ❤️ by huzaifa • v" + version.Version)
}

func (ui *UI) buildLayout() {
    content := container.NewBorder(
        container.NewVBox(ui.statusDisplay.Widget(), ui.toolbar.Widget()),
        ui.footer.Widget(), nil, nil,
        ui.profileList.Widget(),
    )

    ui.window.SetContent(content)
}

func (ui *UI) refresh() {
    cfg, err := config.LoadConfig()
    if err != nil {
        ui.logger.Errorw("Failed to load config", "error", err)
        return
    }

    ui.config = cfg
    ui.profiles = ui.config.GetProfiles()
    ui.profileList.Refresh()
    ui.statusDisplay.Update(ui.gitManager, ui.config)
    // ensure action/dialogs use the latest config
    if ui.toolbar != nil {
        ui.toolbar.UpdateConfig(ui.config)
    }
    ui.selectedItem = -1
    ui.logger.Infow("Refreshed profile list", "count", len(ui.profiles))
}

// Getters for components to access UI state
func (ui *UI) GetProfiles() []*profile.Profile {
	return ui.profiles
}

func (ui *UI) GetSelectedItem() int {
	return ui.selectedItem
}

func (ui *UI) SetSelectedItem(index int) {
	ui.selectedItem = index
}

func (ui *UI) GetConfig() *config.Config {
	return ui.config
}

func (ui *UI) GetGitManager() *git.Manager {
	return ui.gitManager
}

func (ui *UI) GetLogger() *logger.Logger {
	return ui.logger
}

func (ui *UI) GetWindow() fyne.Window {
	return ui.window
}
