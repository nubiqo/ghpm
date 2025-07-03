package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ProfileList struct {
	ui   *UI
	list *widget.List
}

func NewProfileList(ui *UI) *ProfileList {
	pl := &ProfileList{ui: ui}
	pl.createList()
	return pl
}

func (pl *ProfileList) createList() {
	pl.list = widget.NewList(
		func() int {
			return len(pl.ui.GetProfiles())
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
			profiles := pl.ui.GetProfiles()
			if i >= len(profiles) {
				return
			}

			c := o.(*fyne.Container)
			profile := profiles[i]

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

	pl.list.OnSelected = func(id widget.ListItemID) {
		pl.ui.SetSelectedItem(id)
	}

	pl.list.OnUnselected = func(id widget.ListItemID) {
		pl.ui.SetSelectedItem(-1)
	}
}

func (pl *ProfileList) Widget() fyne.CanvasObject {
	return pl.list
}

func (pl *ProfileList) Refresh() {
	pl.list.Refresh()
	if pl.ui.GetSelectedItem() != -1 {
		pl.list.Select(pl.ui.GetSelectedItem())
	} else {
		pl.list.Unselect(-1)
	}
}
