package main

import (
	"encoding/csv"
	"fmt"
	"github.com/spkg/bom"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func (app *Application) LoadCSV(filename string) {

	f, err := os.Open(filename)
	if err != nil {
		//gtk.MessageDialogNew(app.mainWindow, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Could not open file: %s", err.Error())
		log.Printf("Could not open file: %s", err.Error())
		return
	}
	defer f.Close()

	// Original file is in UTF-8 with BOM?!
	r := bom.NewReader(f)
	csvReader := csv.NewReader(r)
	csvReader.Comma = ';'

	bookList := [][]string{}
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			//gtk.MessageDialogNew(app.mainWindow, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Could not read list: %s", err.Error())
			log.Printf("Could not read list: %s", err.Error())
			return
		}

		// Make sure all columns are accounted for
		for len(record) < 6 {
			record = append(record, "")
		}
		bookList = append(bookList, record)
	}

	app.filename = filename
	app.bookList = bookList[1:] // Skip header

	app.UpdateInfo()
}

func (app *Application) writeCSV(f *os.File, includeAll bool) error {

	// Add a stupid UTF-8 BOM, seems like Excel wants it that way
	f.WriteString("\xef\xbb\xbf")
	csvWriter := csv.NewWriter(f)
	csvWriter.Comma = ';'

	err := csvWriter.Write(csvHeader)
	if err != nil {
		return err
	}

	var list [][]string
	if includeAll {
		list = app.bookList
	} else {
		for k := range app.bookList {
			if app.bookList[k][ColumnFound] == "ja" {
				continue
			}
			list = append(list, app.bookList[k])
		}
	}

	err = csvWriter.WriteAll(list)
	return err
}

func (app *Application) Save() {
	f, err := os.OpenFile(app.filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("Could not open file: %s", err.Error())
		return
	}
	defer f.Close()

	err = app.writeCSV(f, true)
	if err != nil {
		log.Printf("Could not write to file: %s", err.Error())
	}
}

func (app *Application) ExportNotFound() (string, error) {
	tmpfile, err := ioutil.TempFile("", fmt.Sprintf("export-%s.*.csv", time.Now().Format("2006-01-02")))
	if err != nil {
		return "", err
	}

	err = app.writeCSV(tmpfile, false)
	if err != nil {
		os.Remove(tmpfile.Name())
		return "", err
	}

	return tmpfile.Name(), nil
}
