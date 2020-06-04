package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/spkg/bom"
)

func (app *Application) LoadCSV(filename string) {
	if app.filehandle != nil {
		app.filehandle.Close()
	}

	var err error
	app.filehandle, err = os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		app.showError("Could not open file: %s", err.Error())
		log.Printf("Could not open file: %s", err.Error())
		return
	}

	// Original file is in UTF-8 with BOM?!
	r := bom.NewReader(app.filehandle)
	csvReader := csv.NewReader(r)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	bookList := [][]string{}
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			app.showError("Could not read list: %s", err.Error())
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

	app.scannedListStore.Clear()
	for _, bookEntry := range app.bookList {
		if bookEntry[ColumnFound] != "ja" {
			continue
		}

		it := app.scannedListStore.Append()
		err := app.scannedListStore.Set(it, []int{0, 1, 2, 3, 4},
			[]interface{}{bookEntry[ColumnBarcode], bookEntry[ColumnTitle], bookEntry[ColumnShelf], bookEntry[ColumnPlacement1], bookEntry[ColumnDate]})
		if err != nil {
			log.Println("err: ", err)
		}
	}
	app.UpdateInfo()
}

func (app *Application) writeCSV(f *os.File, includeAll bool) error {
	_, err := app.filehandle.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = app.filehandle.Truncate(0)
	if err != nil {
		return err
	}

	// Add a stupid UTF-8 BOM, seems like Excel wants it that way
	f.WriteString("\xef\xbb\xbf")
	csvWriter := csv.NewWriter(f)
	csvWriter.Comma = ';'

	err = csvWriter.Write(csvHeader)
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
	var err error
	if app.filehandle == nil {
		return
	}
	app.fileMutex.Lock()
	defer app.fileMutex.Unlock()

	err = app.writeCSV(app.filehandle, true)
	if err != nil {
		app.showError("Could not save information: %s", err)
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
