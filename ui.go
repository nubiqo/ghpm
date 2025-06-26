package main

import (
	"fmt"
	"log"
	"path/filepath"

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
	profileList   *widget.List
	currentStatus *widget.Label
	selectedItem  int // Track selected item
	profiles      []*Profile
}

// NewUI creates a new UI instance
func NewUI(app fyne.App, config *Config) *UI {
	ui := &UI{
		app:          app,
		config:       config,
		gitManager:   NewGitManager(),
		selectedItem: -1,
	}

	ui.window = app.NewWindow("GitHub Profile Manager")
	ui.window.Resize(fyne.NewSize(900, 700))
	ui.window.CenterOnScreen()

	ui.buildUI()
	ui.refresh()

	return ui
}

// Show displays the main window
func (ui *UI) Show() {
	ui.window.ShowAndRun()
}

// buildUI constructs the main user interface
func (ui *UI) buildUI() {
	// Current status
	ui.currentStatus = widget.NewLabel("Current Profile: Loading...")
	ui.currentStatus.TextStyle = fyne.TextStyle{Bold: true}

	statusCard := widget.NewCard("Current Configuration", "", ui.currentStatus)

	// Profile list
	ui.profileList = widget.NewList(
		func() int {
			return len(ui.profiles)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("Profile Name"),
				layout.NewSpacer(),
				widget.NewLabel("Status"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(ui.profiles) {
				return
			}

			c := o.(*fyne.Container)
			profile := ui.profiles[i]

			icon := c.Objects[0].(*widget.Icon)
			nameLabel := c.Objects[1].(*widget.Label)
			statusLabel := c.Objects[3].(*widget.Label)

			nameLabel.SetText(profile.String())

			if profile.IsActive {
				statusLabel.SetText("ACTIVE")
				statusLabel.TextStyle = fyne.TextStyle{Bold: true}
				icon.SetResource(theme.ConfirmIcon())
			} else {
				statusLabel.SetText("")
				statusLabel.TextStyle = fyne.TextStyle{}
				icon.SetResource(theme.AccountIcon())
			}
		},
	)

	// Set up selection callback
	ui.profileList.OnSelected = func(id widget.ListItemID) {
		ui.selectedItem = id
	}

	ui.profileList.OnUnselected = func(id widget.ListItemID) {
		ui.selectedItem = -1
	}

	// Main action buttons
	addBtn := widget.NewButtonWithIcon("Add Profile", theme.ContentAddIcon(), ui.showAddProfileDialog)
	detectBtn := widget.NewButtonWithIcon("Detect Current", theme.SearchIcon(), ui.detectCurrentProfile)

	// Profile management buttons
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), ui.showEditProfileDialog)
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), ui.deleteProfile)
	cloneBtn := widget.NewButtonWithIcon("Clone", theme.ContentCopyIcon(), ui.cloneProfile)
	exportBtn := widget.NewButtonWithIcon("Export", theme.DownloadIcon(), ui.exportProfile)

	// Operation buttons
	switchBtn := widget.NewButtonWithIcon("Switch Profile", theme.ConfirmIcon(), ui.switchProfile)
	testSSHBtn := widget.NewButtonWithIcon("Test SSH", theme.ComputerIcon(), ui.testSSH)
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), ui.refresh)

	// Button layout
	topButtonBar := container.NewHBox(
		addBtn,
		detectBtn,
		widget.NewSeparator(),
		editBtn,
		deleteBtn,
		cloneBtn,
		exportBtn,
	)

	bottomButtonBar := container.NewHBox(
		switchBtn,
		testSSHBtn,
		widget.NewSeparator(),
		refreshBtn,
	)

	buttonContainer := container.NewVBox(topButtonBar, bottomButtonBar)

	// Main layout
	content := container.NewBorder(
		container.NewVBox(statusCard, buttonContainer),
		nil, nil, nil,
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

	// Find matching profile
	active := ui.config.GetActiveProfile()
	if active != nil {
		status := fmt.Sprintf("Profile: %s\nGit: %s <%s>", active.Name, username, email)
		if active.HasSSHKeys() {
			if fingerprint, err := ui.gitManager.GetSSHKeyFingerprint(); err == nil {
				status += fmt.Sprintf("\nSSH: %s", fingerprint)
			} else {
				status += "\nSSH: Available"
			}
		}
		ui.currentStatus.SetText(status)
	} else {
		ui.currentStatus.SetText(fmt.Sprintf("Git: %s <%s>\n(No active profile)", username, email))
	}
}

// detectCurrentProfile detects the current system configuration
func (ui *UI) detectCurrentProfile() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter profile name")

	form := widget.NewForm(
		widget.NewFormItem("Profile Name", nameEntry),
	)

	dlg := dialog.NewCustomConfirm("Detect Current Configuration", "Create", "Cancel", form, func(create bool) {
		if !create || nameEntry.Text == "" {
			return
		}

		profile, err := CreateFromSystem(nameEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to detect configuration: %w", err), ui.window)
			return
		}

		if err := ui.config.AddProfile(profile); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.refresh()
		log.Printf("Created profile from system: %s", profile.Name)

		dialog.ShowInformation("Success",
			fmt.Sprintf("Created profile '%s' from current system configuration", profile.Name),
			ui.window)
	}, ui.window)

	dlg.Resize(fyne.NewSize(400, 200))
	dlg.Show()
}

// showAddProfileDialog shows the dialog to add a new profile
func (ui *UI) showAddProfileDialog() {
	ui.showProfileDialog(nil, "Add Profile", func(p *Profile) {
		if err := ui.config.AddProfile(p); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		ui.refresh()
		log.Printf("Added profile: %s", p.Name)
	})
}

// showEditProfileDialog shows the dialog to edit the selected profile
func (ui *UI) showEditProfileDialog() {
	if ui.selectedItem < 0 || ui.selectedItem >= len(ui.profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to edit", ui.window)
		return
	}

	profile := ui.profiles[ui.selectedItem]
	ui.showProfileDialog(profile, "Edit Profile", func(p *Profile) {
		if err := ui.config.UpdateProfile(profile.Name, p); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		ui.refresh()
		log.Printf("Updated profile: %s", p.Name)
	})
}

// showProfileDialog shows a dialog for adding/editing profiles
func (ui *UI) showProfileDialog(profile *Profile, title string, onSave func(*Profile)) {
	// Create form fields
	nameEntry := widget.NewEntry()
	usernameEntry := widget.NewEntry()
	emailEntry := widget.NewEntry()

	// SSH key display and selection
	privateKeyLabel := widget.NewLabel("No private key")
	publicKeyLabel := widget.NewLabel("No public key")

	var privateKeyContent, publicKeyContent string

	// Pre-fill if editing
	if profile != nil {
		nameEntry.SetText(profile.Name)
		usernameEntry.SetText(profile.GitUsername)
		emailEntry.SetText(profile.GitEmail)
		privateKeyContent = profile.SSHPrivateKey
		publicKeyContent = profile.SSHPublicKey

		if privateKeyContent != "" {
			privateKeyLabel.SetText("Private key loaded")
		}
		if publicKeyContent != "" {
			publicKeyLabel.SetText("Public key loaded")
		}
	}

	// SSH key selection buttons
	selectPrivateBtn := widget.NewButton("Select Private Key", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			data := make([]byte, 10240) // Limit key size
			n, err := reader.Read(data)
			if err != nil && n == 0 {
				dialog.ShowError(err, ui.window)
				return
			}

			keyContent := string(data[:n])
			if err := ui.gitManager.ValidateSSHKey(keyContent, true); err != nil {
				dialog.ShowError(fmt.Errorf("invalid private key: %w", err), ui.window)
				return
			}

			privateKeyContent = keyContent
			privateKeyLabel.SetText(fmt.Sprintf("Loaded: %s", filepath.Base(reader.URI().Path())))
		}, ui.window)
	})

	selectPublicBtn := widget.NewButton("Select Public Key", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			data := make([]byte, 2048) // Public keys are smaller
			n, err := reader.Read(data)
			if err != nil && n == 0 {
				dialog.ShowError(err, ui.window)
				return
			}

			keyContent := string(data[:n])
			if err := ui.gitManager.ValidateSSHKey(keyContent, false); err != nil {
				dialog.ShowError(fmt.Errorf("invalid public key: %w", err), ui.window)
				return
			}

			publicKeyContent = keyContent
			publicKeyLabel.SetText(fmt.Sprintf("Loaded: %s", filepath.Base(reader.URI().Path())))
		}, ui.window)
	})

	clearKeysBtn := widget.NewButton("Clear SSH Keys", func() {
		privateKeyContent = ""
		publicKeyContent = ""
		privateKeyLabel.SetText("No private key")
		publicKeyLabel.SetText("No public key")
	})

	// Create form
	form := widget.NewForm(
		widget.NewFormItem("Profile Name*", nameEntry),
		widget.NewFormItem("Git Username*", usernameEntry),
		widget.NewFormItem("Git Email*", emailEntry),
		widget.NewFormItem("Private Key", container.NewBorder(nil, nil, selectPrivateBtn, nil, privateKeyLabel)),
		widget.NewFormItem("Public Key", container.NewBorder(nil, nil, selectPublicBtn, nil, publicKeyLabel)),
		widget.NewFormItem("", clearKeysBtn),
	)

	// Help text
	helpText := widget.NewLabel("* Required fields\nSSH keys are stored securely in the profile")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(form, helpText)

	// Create dialog
	dlg := dialog.NewCustomConfirm(title, "Save", "Cancel", content, func(save bool) {
		if !save {
			return
		}

		// Create profile
		p := &Profile{
			Name:          nameEntry.Text,
			GitUsername:   usernameEntry.Text,
			GitEmail:      emailEntry.Text,
			SSHPrivateKey: privateKeyContent,
			SSHPublicKey:  publicKeyContent,
			CreatedFrom:   "manual",
		}

		// Validate
		if err := p.Validate(); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		onSave(p)
	}, ui.window)

	dlg.Resize(fyne.NewSize(600, 500))
	dlg.Show()
}

// cloneProfile clones the selected profile
func (ui *UI) cloneProfile() {
	if ui.selectedItem < 0 || ui.selectedItem >= len(ui.profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to clone", ui.window)
		return
	}

	sourceProfile := ui.profiles[ui.selectedItem]

	nameEntry := widget.NewEntry()
	nameEntry.SetText(sourceProfile.Name + "_copy")

	form := widget.NewForm(
		widget.NewFormItem("New Profile Name", nameEntry),
	)

	dlg := dialog.NewCustomConfirm("Clone Profile", "Clone", "Cancel", form, func(clone bool) {
		if !clone || nameEntry.Text == "" {
			return
		}

		newProfile := sourceProfile.Clone(nameEntry.Text)
		if err := ui.config.AddProfile(newProfile); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.refresh()
		log.Printf("Cloned profile: %s -> %s", sourceProfile.Name, newProfile.Name)
	}, ui.window)

	dlg.Resize(fyne.NewSize(400, 200))
	dlg.Show()
}

// exportProfile exports the selected profile
func (ui *UI) exportProfile() {
	if ui.selectedItem < 0 || ui.selectedItem >= len(ui.profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to export", ui.window)
		return
	}

	profile := ui.profiles[ui.selectedItem]

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil || dir == nil {
			return
		}

		exportPath := dir.Path()
		if err := ui.config.ExportProfile(profile.Name, exportPath); err != nil {
			dialog.ShowError(fmt.Errorf("failed to export profile: %w", err), ui.window)
			return
		}

		log.Printf("Exported profile %s to %s", profile.Name, exportPath)
		dialog.ShowInformation("Success",
			fmt.Sprintf("Profile '%s' exported to:\n%s", profile.Name, filepath.Join(exportPath, profile.Name+".json")),
			ui.window)
	}, ui.window)
}

// deleteProfile deletes the selected profile
func (ui *UI) deleteProfile() {
	if ui.selectedItem < 0 || ui.selectedItem >= len(ui.profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to delete", ui.window)
		return
	}

	profile := ui.profiles[ui.selectedItem]

	dialog.ShowConfirm("Delete Profile",
		fmt.Sprintf("Are you sure you want to delete profile '%s'?\nThis action cannot be undone.", profile.Name),
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := ui.config.DeleteProfile(profile.Name); err != nil {
				dialog.ShowError(err, ui.window)
				return
			}

			ui.refresh()
			log.Printf("Deleted profile: %s", profile.Name)
		}, ui.window)
}

// switchProfile switches to the selected profile
func (ui *UI) switchProfile() {
	if ui.selectedItem < 0 || ui.selectedItem >= len(ui.profiles) {
		dialog.ShowInformation("No Selection", "Please select a profile to switch to", ui.window)
		return
	}

	profile := ui.profiles[ui.selectedItem]

	// Show confirmation dialog
	message := fmt.Sprintf("Switch to profile '%s'?\n\nThis will:\n• Set git config to %s <%s>",
		profile.Name, profile.GitUsername, profile.GitEmail)

	if profile.HasSSHKeys() {
		message += "\n• Replace SSH keys with profile keys"
	}

	dialog.ShowConfirm("Switch Profile", message, func(confirm bool) {
		if !confirm {
			return
		}

		// Show progress dialog
		progressDlg := dialog.NewProgressInfinite("Switching Profile", "Configuring git and SSH...", ui.window)
		progressDlg.Show()

		go func() {
			// Switch profile
			err := ui.gitManager.SwitchProfile(profile)

			fyne.DoAndWait(func() {
				progressDlg.Hide()

				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to switch profile: %w", err), ui.window)
					return
				}

				// Update active status
				if err := ui.config.SetActiveProfile(profile.Name); err != nil {
					dialog.ShowError(err, ui.window)
					return
				}

				ui.refresh()

				successMsg := fmt.Sprintf("Switched to profile '%s'", profile.Name)
				if profile.HasSSHKeys() {
					successMsg += "\nSSH keys have been configured"
				}

				dialog.ShowInformation("Success", successMsg, ui.window)
			})
		}()
	}, ui.window)
}

// testSSH tests the SSH connection
func (ui *UI) testSSH() {
	progressDlg := dialog.NewProgressInfinite("Testing SSH", "Testing SSH connection to GitHub...", ui.window)
	progressDlg.Show()

	go func() {
		err := ui.gitManager.TestSSHConnection()

		fyne.DoAndWait(func() {
			progressDlg.Hide()

			if err != nil {
				dialog.ShowError(fmt.Errorf("SSH test failed: %w", err), ui.window)
			} else {
				dialog.ShowInformation("Success", "SSH connection to GitHub successful!", ui.window)
			}
		})
	}()
}

// refresh refreshes the UI
func (ui *UI) refresh() {
	// Reload config
	cfg, err := LoadConfig()
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}

	ui.config = cfg
	ui.profiles = ui.config.GetProfiles()
	ui.profileList.Refresh()
	ui.updateCurrentStatus()
	ui.selectedItem = -1 // Clear selection
	log.Printf("Refreshed profile list - found %d profiles", len(ui.profiles))
}
