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

	app.lblInfo.SetText(fmt.Sprintf("%s: %d exemplar, %d skannade", filepath.Base(app.filename), len(app.bookList), seen))
}
