package main

import (
	"fmt"
	"path/filepath"
)

func (app *Application) UpdateInfo() {
	seen := 0
	for k := range app.bookList {
		if app.bookList[k][ColumnFound] == "ja" {
			seen++
		}
	}

	var percent float32
	if len(app.bookList) > 0 {
		percent = (float32(seen) / float32(len(app.bookList))) * 100
	}

	app.lblInfo.SetText(fmt.Sprintf("%s: %d exemplar, %d skannade (%.2f %%)", filepath.Base(app.filename), len(app.bookList), seen, percent))
}
