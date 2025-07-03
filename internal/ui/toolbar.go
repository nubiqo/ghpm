package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/huzaifanur/ghpm/internal/profile"
	"github.com/huzaifanur/ghpm/internal/ui/actions"
	"github.com/huzaifanur/ghpm/internal/ui/dialogs"
)

type Toolbar struct {
	ui             *UI
	container      *fyne.Container
	profileActions *actions.ProfileActions
	profileDialog  *dialogs.ProfileDialog
	detectDialog   *dialogs.DetectDialog
}

func NewToolbar(ui *UI) *Toolbar {
	tb := &Toolbar{ui: ui}
	tb.initializeComponents()
	tb.createToolbar()
	return tb
}

func (tb *Toolbar) initializeComponents() {
	tb.profileActions = actions.NewProfileActions(
		tb.ui.GetWindow(),
		tb.ui.GetConfig(),
		tb.ui.GetGitManager(),
		tb.ui.GetLogger(),
	)
	tb.profileDialog = dialogs.NewProfileDialog(
		tb.ui.GetWindow(),
		tb.ui.GetGitManager(),
	)
	tb.detectDialog = dialogs.NewDetectDialog(
		tb.ui.GetWindow(),
		tb.ui.GetConfig(),
		tb.ui.GetLogger(),
	)
}

func (tb *Toolbar) createToolbar() {
	// Main action buttons
	addBtn := widget.NewButtonWithIcon("Add Profile", theme.ContentAddIcon(), tb.showAddProfileDialog)
	detectBtn := widget.NewButtonWithIcon("Detect Current", theme.SearchIcon(), tb.detectCurrentProfile)

	// Profile management buttons
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), tb.showEditProfileDialog)
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), tb.deleteProfile)
	importBtn := widget.NewButtonWithIcon("Import", theme.UploadIcon(), tb.importProfile)
	exportBtn := widget.NewButtonWithIcon("Export", theme.DownloadIcon(), tb.exportProfile)

	// Operation buttons
	switchBtn := widget.NewButtonWithIcon("Switch Profile", theme.ConfirmIcon(), tb.switchProfile)
	testSSHBtn := widget.NewButtonWithIcon("Test SSH", theme.ComputerIcon(), tb.testSSH)
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), tb.refresh)

	// Button layout
	topButtonBar := container.NewHBox(
		addBtn,
		detectBtn,
		widget.NewSeparator(),
		editBtn,
		deleteBtn,
		importBtn,
		exportBtn,
	)

	bottomButtonBar := container.NewHBox(
		switchBtn,
		testSSHBtn,
		widget.NewSeparator(),
		refreshBtn,
	)

	tb.container = container.NewVBox(topButtonBar, bottomButtonBar)
}

func (tb *Toolbar) Widget() fyne.CanvasObject {
	return tb.container
}

func (tb *Toolbar) detectCurrentProfile() {
	tb.detectDialog.Show(func() {
		tb.ui.refresh()
	})
}

func (tb *Toolbar) showAddProfileDialog() {
	tb.profileDialog.Show(nil, "Add Profile", func(p *profile.Profile) {
		if err := tb.ui.GetConfig().AddProfile(p); err != nil {
			tb.ui.GetLogger().Errorw("Failed to add profile", "error", err)
			return
		}
		tb.ui.refresh()
		tb.ui.GetLogger().Infow("Added profile", "name", p.Name)
	})
}

func (tb *Toolbar) showEditProfileDialog() {
	selectedProfile := tb.getSelectedProfile()
	if selectedProfile == nil {
		return
	}

	tb.profileDialog.Show(selectedProfile, "Edit Profile", func(p *profile.Profile) {
		if err := tb.ui.GetConfig().UpdateProfile(selectedProfile.Name, p); err != nil {
			tb.ui.GetLogger().Errorw("Failed to update profile", "error", err)
			return
		}
		tb.ui.refresh()
		tb.ui.GetLogger().Infow("Updated profile", "name", p.Name)
	})
}

func (tb *Toolbar) importProfile() {
	tb.profileActions.Import(func() {
		tb.ui.refresh()
	})
}

func (tb *Toolbar) exportProfile() {
	selectedProfile := tb.getSelectedProfile()
	tb.profileActions.Export(selectedProfile)
}

func (tb *Toolbar) deleteProfile() {
	selectedProfile := tb.getSelectedProfile()
	tb.profileActions.Delete(selectedProfile, func() {
		tb.ui.refresh()
	})
}

func (tb *Toolbar) switchProfile() {
	selectedProfile := tb.getSelectedProfile()
	tb.profileActions.Switch(selectedProfile, func() {
		tb.ui.refresh()
	})
}

func (tb *Toolbar) testSSH() {
	tb.profileActions.TestSSH()
}

func (tb *Toolbar) refresh() {
	tb.ui.refresh()
}

func (tb *Toolbar) getSelectedProfile() *profile.Profile {
	selectedItem := tb.ui.GetSelectedItem()
	profiles := tb.ui.GetProfiles()

	if selectedItem < 0 || selectedItem >= len(profiles) {
		return nil
	}

	return profiles[selectedItem]
}
