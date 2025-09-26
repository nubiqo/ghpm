package dialogs

import (
    "fmt"
    "io"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
    "github.com/huzaifanur/ghpm/internal/git"
    "github.com/huzaifanur/ghpm/internal/profile"
)

type ProfileDialog struct {
	window     fyne.Window
	gitManager *git.Manager
}

func NewProfileDialog(window fyne.Window, gitManager *git.Manager) *ProfileDialog {
	return &ProfileDialog{
		window:     window,
		gitManager: gitManager,
	}
}

func (pd *ProfileDialog) Show(editProfile *profile.Profile, title string, onSave func(*profile.Profile)) {
	nameEntry := widget.NewEntry()
	usernameEntry := widget.NewEntry()
	emailEntry := widget.NewEntry()

	privateKeyLabel := widget.NewLabel("No private key")
	privateKeyLabel.Wrapping = fyne.TextWrapWord
	publicKeyLabel := widget.NewLabel("No public key")
	publicKeyLabel.Wrapping = fyne.TextWrapWord

	var privateKeyContent, publicKeyContent string

	if editProfile != nil {
		nameEntry.SetText(editProfile.Name)
		usernameEntry.SetText(editProfile.GitUsername)
		emailEntry.SetText(editProfile.GitEmail)
		privateKeyContent = editProfile.SSHPrivateKey
		publicKeyContent = editProfile.SSHPublicKey

		if privateKeyContent != "" {
			privateKeyLabel.SetText("Private key loaded from profile")
		}
		if publicKeyContent != "" {
			publicKeyLabel.SetText("Public key loaded from profile")
		}
	}

	selectPrivateBtn := widget.NewButton("Select Private Key", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			keyContent, err := pd.readSSHKeyFile(reader, true)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to read private key: %w", err), pd.window)
				return
			}

			if err := pd.gitManager.ValidateSSHKey(keyContent, true); err != nil {
				dialog.ShowError(fmt.Errorf("invalid private key: %w", err), pd.window)
				return
			}

			privateKeyContent = keyContent
			privateKeyLabel.SetText(fmt.Sprintf("Loaded: %s", reader.URI().Path()))
		}, pd.window)
		fileDialog.Resize(fyne.NewSize(800, 600))
		fileDialog.Show()
	})

    selectPublicBtn := widget.NewButton("Select Public Key", func() {
        fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
            if err != nil || reader == nil {
                return
            }
            defer reader.Close()

			keyContent, err := pd.readSSHKeyFile(reader, false)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to read public key: %w", err), pd.window)
				return
			}

			if err := pd.gitManager.ValidateSSHKey(keyContent, false); err != nil {
				dialog.ShowError(fmt.Errorf("invalid public key: %w", err), pd.window)
				return
			}

			publicKeyContent = keyContent
			publicKeyLabel.SetText(fmt.Sprintf("Loaded: %s", reader.URI().Path()))
		}, pd.window)
        fileDialog.Resize(fyne.NewSize(800, 600))
        fileDialog.Show()
    })

    pastePrivateBtn := widget.NewButton("Paste Private Key", func() {
        entry := widget.NewMultiLineEntry()
        entry.SetPlaceHolder("Paste your private key, including BEGIN/END lines")
        dlg := dialog.NewCustomConfirm("Paste Private Key", "Use", "Cancel", entry, func(ok bool) {
            if !ok {
                return
            }
            keyContent := entry.Text
            if err := pd.gitManager.ValidateSSHKey(keyContent, true); err != nil {
                dialog.ShowError(fmt.Errorf("invalid private key: %w", err), pd.window)
                return
            }
            privateKeyContent = keyContent
            privateKeyLabel.SetText("Private key pasted")
        }, pd.window)
        dlg.Resize(fyne.NewSize(700, 500))
        dlg.Show()
    })

    pastePublicBtn := widget.NewButton("Paste Public Key", func() {
        entry := widget.NewMultiLineEntry()
        entry.SetPlaceHolder("Paste your public key (ssh-...)")
        dlg := dialog.NewCustomConfirm("Paste Public Key", "Use", "Cancel", entry, func(ok bool) {
            if !ok {
                return
            }
            keyContent := entry.Text
            if err := pd.gitManager.ValidateSSHKey(keyContent, false); err != nil {
                dialog.ShowError(fmt.Errorf("invalid public key: %w", err), pd.window)
                return
            }
            publicKeyContent = keyContent
            publicKeyLabel.SetText("Public key pasted")
        }, pd.window)
        dlg.Resize(fyne.NewSize(700, 400))
        dlg.Show()
    })

	form := widget.NewForm(
		widget.NewFormItem("Profile Name*", nameEntry),
		widget.NewFormItem("Git Username*", usernameEntry),
		widget.NewFormItem("Git Email*", emailEntry),
	)

    sshContainer := container.NewVBox(
        widget.NewLabel("SSH Keys*"),
        container.NewBorder(nil, nil, container.NewHBox(selectPrivateBtn, pastePrivateBtn), nil, privateKeyLabel),
        container.NewBorder(nil, nil, container.NewHBox(selectPublicBtn, pastePublicBtn), nil, publicKeyLabel),
    )

	helpText := widget.NewLabel("* Required fields")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		sshContainer,
		widget.NewSeparator(),
		helpText,
	)

	dlg := dialog.NewCustomConfirm(title, "Save", "Cancel", content, func(save bool) {
		if !save {
			return
		}

		p := &profile.Profile{
			Name:          nameEntry.Text,
			GitUsername:   usernameEntry.Text,
			GitEmail:      emailEntry.Text,
			SSHPrivateKey: privateKeyContent,
			SSHPublicKey:  publicKeyContent,
			CreatedFrom:   "manual",
		}

		if err := p.Validate(); err != nil {
			dialog.ShowError(err, pd.window)
			return
		}

		onSave(p)
	}, pd.window)

	dlg.Resize(fyne.NewSize(700, 600))
	dlg.Show()
}

func (pd *ProfileDialog) readSSHKeyFile(reader fyne.URIReadCloser, isPrivate bool) (string, error) {
	var maxSize int64
	if isPrivate {
		maxSize = 16384 // 16KB for private keys
	} else {
		maxSize = 4096 // 4KB for public keys
	}

	limitedReader := io.LimitReader(reader, maxSize+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if int64(len(data)) > maxSize {
		return "", fmt.Errorf("file too large (max %d bytes for %s key)",
			maxSize, map[bool]string{true: "private", false: "public"}[isPrivate])
	}

	if len(data) < 10 {
		return "", fmt.Errorf("file too small to be a valid SSH key")
	}

	return string(data), nil
}
