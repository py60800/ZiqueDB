// AbcImport
package main

import (
	"github.com/py60800/zique/zdb"

	"github.com/gotk3/gotk3/gtk"
)

type AbcImporter struct {
}

func (c *ZContext) MkAbcImport() (*AbcImporter, gtk.IWidget) {

	l := &AbcImporter{}

	menuButton, _ := gtk.MenuButtonNew()
	menuButton.SetLabel("AbcImport...")
	popover, _ := gtk.PopoverNew(menuButton)
	menuButton.SetPopover(popover)
	grid, _ := gtk.GridNew()
	grid.SetColumnHomogeneous(true)
	menuButton.Connect("clicked", func() {
		popover.ShowAll()
	})
	popover.Add(grid)
	textView, _ := gtk.TextViewNew()
	textView.SetSizeRequest(100, 200)
	is := 0
	grid.Attach(textView, 0, 0, 6, 8)
	is += 8
	importB, _ := gtk.ButtonNewWithLabel("Import")
	importB.Connect("clicked", func() {
		b, _ := textView.GetBuffer()
		start, end := b.GetBounds()
		txt, _ := b.GetText(start, end, true)
		msg := zdb.AbcImport(txt)
		if msg != "" {
			Message(msg)
		}
	})
	clearB, _ := gtk.ButtonNewWithLabel("Clear")
	clearB.Connect("clicked", func() {
		b, _ := textView.GetBuffer()
		b.SetText("")
	})
	cancelB, _ := gtk.ButtonNewWithLabel("Cancel")
	cancelB.Connect("clicked", func() {
		popover.Popdown()
	})
	grid.Attach(importB, 0, is, 2, 1)
	grid.Attach(clearB, 2, is, 2, 1)
	grid.Attach(cancelB, 4, is, 2, 1)

	return l, menuButton
}
