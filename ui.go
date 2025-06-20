package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UI represents the main user interface
type UI struct {
	app        fyne.App
	window     fyne.Window
	config     *Config
	gitManager *GitManager

	// UI components
	profileList     *widget.List
	currentStatus   *widget.Label
	selectedProfile int
}

// NewUI creates a new UI instance
func NewUI(app fyne.App, config *Config) *UI {
	ui := &UI{
		app:             app,
		config:          config,
		gitManager:      NewGitManager(),
		selectedProfile: -1,
	}

	ui.window = app.NewWindow("GitHub Profile Manager")
	ui.window.Resize(fyne.NewSize(800, 600))
	ui.window.CenterOnScreen()

	ui.buildUI()
	ui.updateCurrentStatus()

	return ui
}

// Show displays the main window
func (ui *UI) Show() {
	ui.window.ShowAndRun()
}

// buildUI constructs the main user interface
func (ui *UI) buildUI() {
	ui.currentStatus = widget.NewLabel("Current Profile: None")
	ui.currentStatus.TextStyle = fyne.TextStyle{Bold: true}

	ui.profileList = widget.NewList(
		func() int {
			return len(ui.config.Profiles)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Profile"),
				layout.NewSpacer(),
				widget.NewLabel("Status"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			c := o.(*fyne.Container)
			profile := ui.config.Profiles[i]

			nameLabel := c.Objects[0].(*widget.Label)
			statusLabel := c.Objects[2].(*widget.Label)

			nameLabel.SetText(fmt.Sprintf("%s - %s", profile.Name, profile.String()))
			if profile.IsActive {
				statusLabel.SetText("ACTIVE")
				statusLabel.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				statusLabel.SetText("")
				statusLabel.TextStyle = fyne.TextStyle{}
			}
		},
	)

	ui.profileList.OnSelected = func(id widget.ListItemID) {
		ui.selectedProfile = id
	}

	ui.profileList.OnUnselected = func(id widget.ListItemID) {
		ui.selectedProfile = -1
	}

	addBtn := widget.NewButtonWithIcon("Add Profile", theme.ContentAddIcon(), ui.showAddProfileDialog)
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), ui.showEditProfileDialog)
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), ui.deleteProfile)
	switchBtn := widget.NewButtonWithIcon("Switch Profile", theme.ConfirmIcon(), ui.switchProfile)
	testSSHBtn := widget.NewButtonWithIcon("Test SSH", theme.ComputerIcon(), ui.testSSH)
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), ui.refresh)

	buttonBar := container.NewHBox(
		addBtn,
		editBtn,
		deleteBtn,
		widget.NewSeparator(),
		switchBtn,
		testSSHBtn,
		widget.NewSeparator(),
		refreshBtn,
	)

	statusCard := container.NewVBox(ui.currentStatus)
	content := container.NewBorder(
		statusCard,
		buttonBar,
		nil,
		nil,
		ui.profileList,
	)

	ui.window.SetContent(content)
}

// updateCurrentStatus updates the current profile status
func (ui *UI) updateCurrentStatus() {
	username, email, err := ui.gitManager.GetCurrentGitConfig()
	if err != nil {
		ui.currentStatus.SetText("Current Profile: Error reading git config")
		return
	}

	active := ui.config.GetActiveProfile()
	if active != nil {
		ui.currentStatus.SetText(fmt.Sprintf("Current Profile: %s (%s <%s>)", active.Name, username, email))
	} else {
		ui.currentStatus.SetText(fmt.Sprintf("Current Profile: %s <%s> (no profile)", username, email))
	}
}

// showAddProfileDialog shows the dialog to add a new profile
func (ui *UI) showAddProfileDialog() {
	ui.showProfileDialog(nil, "Add Profile", func(p Profile) {
		if err := ui.config.AddProfile(p); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		ui.profileList.Refresh()
		ui.updateCurrentStatus()
		log.Printf("Added profile: %s", p.Name)
	})
}

// showEditProfileDialog shows the dialog to edit the selected profile
func (ui *UI) showEditProfileDialog() {
	if ui.selectedProfile < 0 || ui.selectedProfile >= len(ui.config.Profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to edit", ui.window)
		return
	}

	profile := ui.config.Profiles[ui.selectedProfile]
	ui.showProfileDialog(&profile, "Edit Profile", func(p Profile) {
		if err := ui.config.UpdateProfile(profile.Name, p); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		ui.profileList.Refresh()
		ui.updateCurrentStatus()
		log.Printf("Updated profile: %s", p.Name)
	})
}

// showProfileDialog shows a dialog for adding/editing profiles
func (ui *UI) showProfileDialog(profile *Profile, title string, onSave func(Profile)) {
	nameEntry := widget.NewEntry()
	usernameEntry := widget.NewEntry()
	emailEntry := widget.NewEntry()
	privateKeyEntry := widget.NewEntry()
	publicKeyEntry := widget.NewEntry()

	if profile != nil {
		nameEntry.SetText(profile.Name)
		usernameEntry.SetText(profile.GitUsername)
		emailEntry.SetText(profile.GitEmail)
		privateKeyEntry.SetText(profile.SSHPrivateKey)
		publicKeyEntry.SetText(profile.SSHPublicKey)
	}

	form := widget.NewForm(
		widget.NewFormItem("Profile Name*", nameEntry),
		widget.NewFormItem("Git Username*", usernameEntry),
		widget.NewFormItem("Git Email*", emailEntry),
		widget.NewFormItem("SSH Private Key Path", privateKeyEntry),
		widget.NewFormItem("SSH Public Key Path", publicKeyEntry),
	)

	helpText := widget.NewLabel("* Required fields\nSSH key paths support $HOME expansion")
	helpText.TextStyle = fyne.TextStyle{Italic: true}
	content := container.NewVBox(form, helpText)

	dlg := dialog.NewCustomConfirm(title, "Save", "Cancel", content, func(save bool) {
		if !save {
			return
		}

		p := Profile{
			Name:          nameEntry.Text,
			GitUsername:   usernameEntry.Text,
			GitEmail:      emailEntry.Text,
			SSHPrivateKey: privateKeyEntry.Text,
			SSHPublicKey:  publicKeyEntry.Text,
		}

		if err := p.Validate(); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		onSave(p)
	}, ui.window)
	dlg.Show()
}

// deleteProfile deletes the selected profile
func (ui *UI) deleteProfile() {
	if ui.selectedProfile < 0 || ui.selectedProfile >= len(ui.config.Profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to delete", ui.window)
		return
	}

	profile := ui.config.Profiles[ui.selectedProfile]
	dialog.ShowConfirm("Delete Profile",
		fmt.Sprintf("Are you sure you want to delete profile '%s'?", profile.Name),
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := ui.config.DeleteProfile(profile.Name); err != nil {
				dialog.ShowError(err, ui.window)
				return
			}

			ui.selectedProfile = -1
			ui.profileList.Refresh()
			ui.updateCurrentStatus()
		}, ui.window)
}

// switchProfile switches to the selected profile
func (ui *UI) switchProfile() {
	if ui.selectedProfile < 0 || ui.selectedProfile >= len(ui.config.Profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to switch to", ui.window)
		return
	}

	profile := &ui.config.Profiles[ui.selectedProfile]

	message := fmt.Sprintf("Switch to profile '%s'?\n\nThis will:\n- Set git config to %s <%s>",
		profile.Name, profile.GitUsername, profile.GitEmail)

	if profile.SSHPrivateKey != "" || profile.SSHPublicKey != "" {
		message += "\n- Replace SSH keys (existing keys will be backed up)"
	}

	dialog.ShowConfirm("Switch Profile", message, func(confirm bool) {
		if !confirm {
			return
		}

		if err := ui.gitManager.SwitchProfile(profile); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		if err := ui.config.SetActiveProfile(profile.Name); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.profileList.Refresh()
		ui.updateCurrentStatus()

		dialog.ShowInformation("Success",
			fmt.Sprintf("Switched to profile '%s'", profile.Name),
			ui.window)
	}, ui.window)
}

// testSSH tests the SSH connection
func (ui *UI) testSSH() {
	progressDlg := dialog.NewProgressInfinite("Testing SSH", "Testing SSH connection to GitHub...", ui.window)
	progressDlg.Show()

	go func() {
		err := ui.gitManager.TestSSHConnection()
		progressDlg.Hide()

		if err != nil {
			dialog.ShowError(fmt.Errorf("SSH test failed: %w", err), ui.window)
		} else {
			dialog.ShowInformation("Success", "SSH connection to GitHub successful!", ui.window)
		}
	}()
}

// refresh refreshes the UI
func (ui *UI) refresh() {
	cfg, err := LoadConfig()
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}

	ui.config = cfg
	ui.profileList.Refresh()
	ui.updateCurrentStatus()
	log.Printf("Refreshed profile list")
}
