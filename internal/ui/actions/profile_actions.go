package actions

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/huzaifanur/ghpm/internal/config"
	"github.com/huzaifanur/ghpm/internal/git"
	"github.com/huzaifanur/ghpm/internal/profile"
	"github.com/huzaifanur/ghpm/pkg/logger"
)

type ProfileActions struct {
	window     fyne.Window
	config     *config.Config
	gitManager *git.Manager
	logger     *logger.Logger
}

func NewProfileActions(window fyne.Window, config *config.Config, gitManager *git.Manager, logger *logger.Logger) *ProfileActions {
	return &ProfileActions{
		window:     window,
		config:     config,
		gitManager: gitManager,
		logger:     logger,
	}
}

func (pa *ProfileActions) SetConfig(cfg *config.Config) {
	pa.config = cfg
}

func (pa *ProfileActions) Import(onComplete func()) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()

		importedProfile, err := pa.config.ImportProfile(filePath)
		if err != nil {
			dialog.ShowError(err, pa.window)
			return
		}

		onComplete()
		pa.logger.Infow("Imported profile", "name", importedProfile.Name, "path", filePath)

		dialog.ShowInformation("Success",
			fmt.Sprintf("Successfully imported profile '%s'", importedProfile.Name),
			pa.window)
	}, pa.window)
	fileDialog.Resize(fyne.NewSize(800, 600))
	fileDialog.Show()
}

func (pa *ProfileActions) Export(selectedProfile *profile.Profile) {
	if selectedProfile == nil {
		dialog.ShowInformation("No Selection", "Please select a profile to export", pa.window)
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil || dir == nil {
			return
		}

		exportPath := dir.Path()
		if err := pa.config.ExportProfile(selectedProfile.Name, exportPath); err != nil {
			dialog.ShowError(fmt.Errorf("failed to export profile: %w", err), pa.window)
			return
		}

		pa.logger.Infow("Exported profile", "name", selectedProfile.Name, "path", exportPath)
		dialog.ShowInformation("Success",
			fmt.Sprintf("Profile '%s' exported to:\n%s", selectedProfile.Name, filepath.Join(exportPath, selectedProfile.Name+".json")),
			pa.window)
	}, pa.window)
}

func (pa *ProfileActions) Delete(selectedProfile *profile.Profile, onComplete func()) {
	if selectedProfile == nil {
		dialog.ShowInformation("No Selection", "Please select a profile to delete", pa.window)
		return
	}

	dialog.ShowConfirm("Delete Profile",
		fmt.Sprintf("Are you sure you want to delete profile '%s'?\nThis action cannot be undone.", selectedProfile.Name),
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := pa.config.DeleteProfile(selectedProfile.Name); err != nil {
				dialog.ShowError(err, pa.window)
				return
			}

			onComplete()
			pa.logger.Infow("Deleted profile", "name", selectedProfile.Name)
		}, pa.window)
}

func (pa *ProfileActions) Switch(selectedProfile *profile.Profile, onComplete func()) {
	if selectedProfile == nil {
		dialog.ShowInformation("No Selection", "Please select a profile to switch to", pa.window)
		return
	}

	message := fmt.Sprintf("Switch to profile '%s'?\n\nThis will:\n• Set git config to %s <%s>",
		selectedProfile.Name, selectedProfile.GitUsername, selectedProfile.GitEmail)

	if selectedProfile.HasSSHKeys() {
		message += "\n• Replace SSH keys with profile keys"
	}

	dialog.ShowConfirm("Switch Profile", message, func(confirm bool) {
		if !confirm {
			return
		}

		progressDlg := dialog.NewProgressInfinite("Switching Profile", "Configuring git and SSH...", pa.window)
		progressDlg.Show()

		go func() {
			err := pa.gitManager.SwitchProfile(selectedProfile)

			fyne.DoAndWait(func() {
				progressDlg.Hide()

				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to switch profile: %w", err), pa.window)
					return
				}

				if err := pa.config.SetActiveProfile(selectedProfile.Name); err != nil {
					dialog.ShowError(err, pa.window)
					return
				}

				onComplete()

				successMsg := fmt.Sprintf("Switched to profile '%s'", selectedProfile.Name)
				if selectedProfile.HasSSHKeys() {
					successMsg += "\nSSH keys have been configured"
				}

				dialog.ShowInformation("Success", successMsg, pa.window)
			})
		}()
	}, pa.window)
}

func (pa *ProfileActions) TestSSH() {
	progressDlg := dialog.NewProgressInfinite("Testing SSH", "Testing SSH connection to GitHub...", pa.window)
	progressDlg.Show()

	go func() {
		err := pa.gitManager.TestSSHConnection()

		fyne.DoAndWait(func() {
			progressDlg.Hide()

			if err != nil {
				dialog.ShowError(fmt.Errorf("SSH test failed: %w", err), pa.window)
			} else {
				dialog.ShowInformation("Success", "SSH connection to GitHub successful!", pa.window)
			}
		})
	}()
}
