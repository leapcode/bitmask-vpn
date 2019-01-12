package dialog

import (
	"os"
	"path/filepath"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func closeDialog(dlg *gtk.Dialog) {
	dlg.Destroy()
}

func (b *MsgBuilder) yesNo() bool {
	ch := make(chan bool)
	_, err := glib.IdleAdd(b._yesNo, ch)
	if err != nil {
		return false
	}
	return <-ch
}

func (b *MsgBuilder) _yesNo(ch chan bool) bool {
	dlg := gtk.MessageDialogNew(nil, 0, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "%s", b.Msg)
	dlg.SetTitle(firstOf(b.Dlg.Title, "Confirm?"))
	b.setIcon(dlg)
	defer closeDialog(&dlg.Dialog)
	ch <- dlg.Run() == gtk.RESPONSE_YES
	return false
}

func (b *MsgBuilder) info() {
	ch := make(chan bool)
	glib.IdleAdd(b._info, ch)
	<-ch
}

func (b *MsgBuilder) _info(ch chan bool) {
	dlg := gtk.MessageDialogNew(nil, 0, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", b.Msg)
	dlg.SetTitle(firstOf(b.Dlg.Title, "Information"))
	b.setIcon(dlg)
	defer closeDialog(&dlg.Dialog)
	dlg.Run()
	ch <- true
}

func (b *MsgBuilder) error() {
	ch := make(chan bool)
	glib.IdleAdd(b._info, ch)
	<-ch
}

func (b *MsgBuilder) _error() {
	dlg := gtk.MessageDialogNew(nil, 0, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", b.Msg)
	dlg.SetTitle(firstOf(b.Dlg.Title, "Error"))
	b.setIcon(dlg)
	defer closeDialog(&dlg.Dialog)
	dlg.Run()
}

func (b *MsgBuilder) setIcon(dlg *gtk.MessageDialog) {
	if b.IconPath == "" {
		return
	}

	pixbuf, err := gdk.PixbufNewFromFile(b.IconPath)
	if err != nil {
		return
	}
	dlg.SetIcon(pixbuf)
}

func (b *FileBuilder) load() (string, error) {
	return chooseFile("Load", gtk.FILE_CHOOSER_ACTION_OPEN, b)
}

func (b *FileBuilder) save() (string, error) {
	f, err := chooseFile("Save", gtk.FILE_CHOOSER_ACTION_SAVE, b)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(f)
	if !os.IsNotExist(err) && !Message("%s already exists, overwrite?", filepath.Base(f)).yesNo() {
		return "", Cancelled
	}
	return f, nil
}

func chooseFile(title string, action gtk.FileChooserAction, b *FileBuilder) (string, error) {
	dlg, err := gtk.FileChooserDialogNewWith2Buttons(firstOf(b.Dlg.Title, title), nil, action, "Ok", gtk.RESPONSE_ACCEPT, "Cancel", gtk.RESPONSE_CANCEL)
	if err != nil {
		return "", err
	}

	for _, filt := range b.Filters {
		filter, err := gtk.FileFilterNew()
		if err != nil {
			return "", err
		}

		filter.SetName(filt.Desc)
		for _, ext := range filt.Extensions {
			filter.AddPattern("*." + ext)
		}
		dlg.AddFilter(filter)
	}
	if b.StartDir != "" {
		dlg.SetCurrentFolder(b.StartDir)
	}
	r := dlg.Run()
	defer closeDialog(&dlg.Dialog)
	if r == gtk.RESPONSE_ACCEPT {
		return dlg.GetFilename(), nil
	}
	return "", Cancelled
}

func (b *DirectoryBuilder) browse() (string, error) {
	return chooseFile("Open Directory", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER, &FileBuilder{Dlg: b.Dlg})
}
