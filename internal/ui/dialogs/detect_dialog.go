package dialogs

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/huzaifanur/ghpm/internal/config"
	"github.com/huzaifanur/ghpm/internal/profile"
	"github.com/huzaifanur/ghpm/pkg/logger"
)

type DetectDialog struct {
	window fyne.Window
	config *config.Config
	logger *logger.Logger
}

func NewDetectDialog(window fyne.Window, config *config.Config, logger *logger.Logger) *DetectDialog {
	return &DetectDialog{
		window: window,
		config: config,
		logger: logger,
	}
}

func (dd *DetectDialog) Show(onProfileCreated func()) {
	detectedProfile, err := profile.CreateFromSystem("temp")
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to detect configuration: %w", err), dd.window)
		return
	}

	previewText := fmt.Sprintf("Git Configuration:\n• Username: %s\n• Email: %s\n\n", detectedProfile.GitUsername, detectedProfile.GitEmail)

	if detectedProfile.HasSSHKeys() {
		previewText += "SSH Keys:\n• Private Key: Found\n• Public Key: Found\n\n"
	} else {
		previewText += "SSH Keys:\n• No SSH keys found\n\n"
	}

	previewText += "This configuration will be saved to the profile."

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter profile name")

	previewLabel := widget.NewLabel(previewText)
	previewLabel.Wrapping = fyne.TextWrapWord

	form := container.NewVBox(
		widget.NewForm(widget.NewFormItem("Profile Name", nameEntry)),
		widget.NewCard("Preview", "", previewLabel),
	)

	dlg := dialog.NewCustomConfirm("Detect Current Configuration", "Create", "Cancel", form, func(create bool) {
		if !create || nameEntry.Text == "" {
			return
		}

		if _, err := dd.config.GetProfile(nameEntry.Text); err == nil {
			dialog.ShowError(fmt.Errorf("profile with name '%s' already exists", nameEntry.Text), dd.window)
			return
		}

		detectedProfile.Name = nameEntry.Text
		if err := dd.config.AddProfile(detectedProfile); err != nil {
			dialog.ShowError(err, dd.window)
			return
		}

		onProfileCreated()
		dd.logger.Infow("Created profile from system", "name", detectedProfile.Name)

		dialog.ShowInformation("Success",
			fmt.Sprintf("Created profile '%s' from current system configuration", detectedProfile.Name),
			dd.window)
	}, dd.window)

	dlg.Resize(fyne.NewSize(600, 500))
	dlg.Show()
}
