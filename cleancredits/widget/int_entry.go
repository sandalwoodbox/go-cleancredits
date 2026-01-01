package widget

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/utils"
)

type IntEntry struct {
	Min, Max int
	widget.Entry
}

func NewIntEntry(min, max int) *IntEntry {
	entry := &IntEntry{
		Min: min,
		Max: max,
	}
	entry.ExtendBaseWidget(entry)
	entry.Scroll = fyne.ScrollNone
	return entry
}

func NewIntEntryWithData(min, max int, data binding.Int) *IntEntry {
	e := NewIntEntry(min, max)
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

func (e *IntEntry) Scrolled(event *fyne.ScrollEvent) {
	val, err := strconv.Atoi(e.Text)
	if err != nil {
		fmt.Println("Error parsing text to int: ", err)
	}
	e.SetText(strconv.Itoa(utils.ClampInt(val+int(event.Scrolled.DY), e.Min, e.Max)))
}

func (e *IntEntry) TypedKey(key *fyne.KeyEvent) {
	val, err := strconv.Atoi(e.Text)
	if err != nil {
		fmt.Println("Error parsing text to int: ", err)
	}
	switch key.Name {
	case fyne.KeyUp:
		if val < e.Max {
			e.SetText(strconv.Itoa(val + 1))
		}
		return
	case fyne.KeyDown:
		if val > e.Min {
			e.SetText(strconv.Itoa(val - 1))
		}
		return
	}
	e.Entry.TypedKey(key)
}
