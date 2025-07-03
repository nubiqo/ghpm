package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/huzaifanur/ghpm/internal/config"
	"github.com/huzaifanur/ghpm/internal/git"
)

type StatusDisplay struct {
	card   *widget.Card
	status *widget.Label
}

func NewStatusDisplay() *StatusDisplay {
	sd := &StatusDisplay{}
	sd.createStatus()
	return sd
}

func (sd *StatusDisplay) createStatus() {
	sd.status = widget.NewLabel("Current Profile: Loading...")
	sd.status.TextStyle = fyne.TextStyle{Bold: true}
	sd.card = widget.NewCard("Current Configuration", "", sd.status)
}

func (sd *StatusDisplay) Widget() fyne.CanvasObject {
	return sd.card
}

func (sd *StatusDisplay) Update(gitManager *git.Manager, cfg *config.Config) {
	username, email, err := gitManager.GetCurrentGitConfig()
	if err != nil {
		sd.status.SetText("Current Profile: Error reading git config")
		return
	}

	active := cfg.GetActiveProfile()
	if active != nil {
		status := fmt.Sprintf("Profile: %s\nGit: %s <%s>", active.Name, username, email)
		if active.HasSSHKeys() {
			if fingerprint, err := gitManager.GetSSHKeyFingerprint(); err == nil {
				status += fmt.Sprintf("\nSSH: %s", fingerprint)
			} else {
				status += "\nSSH: Available"
			}
		}
		sd.status.SetText(status)
	} else {
		sd.status.SetText(fmt.Sprintf("Git: %s <%s>\n(No active profile)", username, email))
	}
}
