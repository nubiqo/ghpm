package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Footer struct {
	container *fyne.Container
}

func NewFooter(attribution string) *Footer {
	label := widget.NewLabel(attribution)
	label.Wrapping = fyne.TextWrapOff
	// right-align using spacer
	c := container.NewHBox(layout.NewSpacer(), label)
	return &Footer{container: c}
}

func (f *Footer) Widget() fyne.CanvasObject { return f.container }
