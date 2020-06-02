package main

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/glib"
)

func (app *Application) AddBook(barcode string) {
	if barcode == "" {
		return
	}

	found := false
	var bookEntry []string
	for k := range app.bookList {
		if app.bookList[k][ColumnBarcode] == barcode {
			found = true
			bookEntry = app.bookList[k]
			break
		}
	}

	if !found {
		fmt.Println("Book not found for barcode", barcode)
		app.lblNotFound.SetText(fmt.Sprintf("%s ej lokaliserat", barcode))
		app.dlgNotFound.ShowAll()
		return
	}

	if bookEntry[ColumnFound] == "ja" {
		app.dlgAlreadySeen.ShowAll()
		glib.TimeoutAdd(2000, app.dlgAlreadySeen.Hide)
		return
	}

	// Mark entry as found
	bookEntry[ColumnFound] = "ja"

	if bookEntry[ColumnDate] != "" {
		app.dlgExpiryDate.ShowAll()
		glib.TimeoutAdd(2000, app.dlgExpiryDate.Hide)
	}

	it := app.scannedListStore.Append()
	err := app.scannedListStore.Set(it,
		[]int{0, 1, 2, 3, 4},
		[]interface{}{barcode, bookEntry[ColumnTitle], bookEntry[ColumnShelf], bookEntry[ColumnPlacement1], bookEntry[ColumnDate]})
	if err != nil {
		log.Println("err: ", err)
	}

	app.Save()
	app.UpdateInfo()
}
