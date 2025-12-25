package widgets

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

type IntEntry struct {
	widget.Entry
}

func NewIntEntry() *IntEntry {
	entry := &IntEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func NewIntEntryWithData(data binding.Int) *IntEntry {
	e := NewIntEntry()
	e.Bind(binding.IntToString(data))
	e.Validator = nil

	return e
}

func (e *IntEntry) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
	}
}

func (e *IntEntry) TypedShortcut(shortcut fyne.Shortcut) {
	paste, ok := shortcut.(*fyne.ShortcutPaste)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}

	content := paste.Clipboard.Content()
	if _, err := strconv.Atoi(content); err == nil {
		e.Entry.TypedShortcut(shortcut)
	}
}

func (e *IntEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}
