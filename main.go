package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/skratchdot/open-golang/open"
)

const appID = "com.yzzyx.inv"

const cssData = `
.bg-green { background-color: #4E9A06; }
.bg-blue { background-color: #729FCF; }
.bg-red { background-color: #CC0000; }
`

type Application struct {
	application      *gtk.Application
	builder          *gtk.Builder
	mainWindow       *gtk.Window
	dlgNotFound      *gtk.Window
	dlgExpiryDate    *gtk.Window
	dlgAlreadySeen   *gtk.Window
	lblNotFound      *gtk.Label
	lblInfo          *gtk.Label
	bookList         [][]string
	scannedListStore *gtk.ListStore
	scannedListView  *gtk.TreeView

	filename   string
	filehandle *os.File
	fileMutex  sync.Mutex
}

var csvHeader = []string{
	"Nr",
	"Huvuduppslag",
	"Etikettnr.",
	"Lån",
	"Avdelning",
	"Placering",
	"Cirk. kat.",
	"Hylla",
	"Annan lånetid",
	"Giltig t.o.m",
	"Avdelning",
	"Placering",
	"Cirk. kat.",
	"Hylla",
	"Status",
	"Lok.dat.",
	"Omlok.dat.",
	"Förf. datum",
	"Publ. nr.",
	"Extern anm.",
	"Anmärkning",
	"Trans.dat.",
	"Id",
	"Namn",
	"Kategori",
	"Grupp",
	"Enhet/Skola",
	"Klass",
	"Extra 1",
	"Personnummer",
	"Adress",
	"Skannad",
}

const (
	ColumnNumber = iota
	ColumnTitle
	ColumnBarcode
	ColumnLoan
	ColumnDept1
	ColumnPlacement1
	ColumnCirkCat1
	ColumnShelf1
	ColumnOtherLoanTime
	ColumnValidUntil
	ColumnDept2
	ColumnPlacement2
	ColumnCirkCat2
	ColumnShelf
	ColumnStatus
	ColumnLocDate
	ColumnReLocDate
	ColumnDate
	ColumnPublNum
	ColumnExternComment
	ColumnComment
	ColumnTransDate
	ColumnId
	ColumnName
	ColumnCategory
	ColumnGroup
	ColumnSchool
	ColumnClass
	ColumnExtra1
	ColumnSSN
	ColumnAddress
	ColumnFound
)

// Add a column to the tree view (during the initialization of the tree view)
func createColumn(title string, id int, toggle bool) *gtk.TreeViewColumn {

	var cellRenderer gtk.ICellRenderer
	var err error
	if !toggle {
		cellRenderer, err = gtk.CellRendererTextNew()
	} else {
		cellRenderer, err = gtk.CellRendererToggleNew()
	}

	if err != nil {
		log.Fatal("Unable to create text cell renderer:", err)
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		log.Fatal("Unable to create cell column:", err)
	}

	return column
}

func (app *Application) showError(fmt string, args ...interface{}) {
	dlg := gtk.MessageDialogNew(app.mainWindow, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, fmt, args...)
	if dlg.Run() == gtk.RESPONSE_OK {
		dlg.Close()
		dlg.Destroy()
	}
}

func (app *Application) menuOpen() {
	fileOpen, err := gtk.FileChooserNativeDialogNew("Välj fil",
		app.mainWindow,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"Ok",
		"Avbryt")
	if err != nil {
		log.Println("Cannot create native file chooser dialog:", err)
		return
	}

	filter, err := gtk.FileFilterNew()
	if err != nil {
		log.Println("Cannot create file filter:", err)
		return
	}
	filter.AddPattern("*.csv")
	filter.SetName("CSV-filer")
	fileOpen.AddFilter(filter)

	result := fileOpen.Run()
	if gtk.ResponseType(result) == gtk.RESPONSE_ACCEPT {
		log.Println("File selected:", fileOpen.GetFilename())
		app.LoadCSV(fileOpen.GetFilename())
	}
}

func (app *Application) menuQuit() {
	app.application.Quit()
}

func (app *Application) keyPress(widget *gtk.Entry, ev *gdk.Event) {
	keyEvent := &gdk.EventKey{ev}

	kv := keyEvent.KeyVal()
	if kv == gdk.KEY_KP_Enter ||
		kv == gdk.KEY_Return {
		text, _ := widget.GetText()

		app.AddBook(text)
		widget.SetText("")
	}
}

func (app *Application) dlgNotFoundBtnPress(widget *gtk.Window, ev *gdk.Event) {
	btnEvent := &gdk.EventButton{ev}

	if btnEvent.Type() == gdk.EVENT_BUTTON_PRESS {
		widget.Hide()
	}
}

func (app *Application) dlgNotFoundKeyPress(widget *gtk.Window, ev *gdk.Event) {
	keyEvent := &gdk.EventKey{ev}

	kv := keyEvent.KeyVal()
	if kv == gdk.KEY_space {
		widget.Hide()
	}
}

func (app *Application) scannedBooksKeyPress(tv *gtk.TreeView, ev *gdk.Event) {
	keyEvent := &gdk.EventKey{ev}

	if keyEvent.KeyVal() != gdk.KEY_c || keyEvent.State()&gdk.GDK_CONTROL_MASK == 0 {
		return
	}

	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err != nil {
		log.Printf("Cannot get clipboard: %s", err)
		return
	}

	selection, err := app.scannedListView.GetSelection()
	if err != nil {
		log.Printf("Cannot get selection: %s", err)
		return
	}

	rows := selection.GetSelectedRows(app.scannedListStore)
	items := make([]string, 0, rows.Length())

	columns := app.scannedListStore.GetNColumns()
	for l := rows; l != nil; l = l.Next() {
		path := l.Data().(*gtk.TreePath)
		iter, _ := app.scannedListStore.GetIter(path)

		values := make([]string, 0, columns)
		for i := 0; i < columns; i++ {
			value, _ := app.scannedListStore.GetValue(iter, i)
			str, _ := value.GetString()
			values = append(values, str)
		}
		// Join columns with tabs, so that they can be copied to excel
		items = append(items, strings.Join(values, "\t"))
	}
	clipboard.SetText(strings.Join(items, "\n"))
}

func (app *Application) btnShowClicked(widget *gtk.Button) {
	if app.filename == "" {
		return
	}

	filename, err := app.ExportNotFound()
	if err != nil {
		log.Printf("Could not write to file: %s", err.Error())
		return
	}

	err = open.Start(filename)
	if err != nil {
		log.Println("cannot start application:", err)
	}
}

func (app *Application) builderFunc() {
	var err error
	app.builder, err = gtk.BuilderNewFromFile("inv.glade")
	//builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatalln("Couldn't make builder:", err)
	}

	obj, err := app.builder.GetObject("dlgMain")
	if err != nil {
		log.Fatalln("Could not get main dialog:", err)
	}
	wnd := obj.(*gtk.Window)

	obj, err = app.builder.GetObject("dlgNotFound")
	if err != nil {
		log.Fatalln("Could not get not-found dialog:", err)
	}
	app.dlgNotFound = obj.(*gtk.Window)

	obj, err = app.builder.GetObject("lblNotFound")
	if err != nil {
		log.Fatalln("Could not get notFound label:", err)
	}
	app.lblNotFound = obj.(*gtk.Label)

	obj, err = app.builder.GetObject("dlgExpiryDate")
	if err != nil {
		log.Fatalln("Could not get expiry dialog:", err)
	}
	app.dlgExpiryDate = obj.(*gtk.Window)

	obj, err = app.builder.GetObject("dlgAlreadySeen")
	if err != nil {
		log.Fatalln("Could not get already seen dialog:", err)
	}
	app.dlgAlreadySeen = obj.(*gtk.Window)

	obj, err = app.builder.GetObject("lblInfo")
	if err != nil {
		log.Fatalln("Could not get info label:", err)
	}
	app.lblInfo = obj.(*gtk.Label)

	obj, err = app.builder.GetObject("scannedBooksTreeView")
	if err != nil {
		log.Fatalln("Could not get tree view:", err)
	}
	tv := obj.(*gtk.TreeView)
	tv.AppendColumn(createColumn("Etikett", 0, false))
	tv.AppendColumn(createColumn("Huvuduppslag", 1, false))
	tv.AppendColumn(createColumn("Hylla", 2, false))
	tv.AppendColumn(createColumn("Placering", 3, false))
	tv.AppendColumn(createColumn("Förfallodatum", 4, false))
	app.scannedListView = tv

	selection, err := tv.GetSelection()
	if err != nil {
		log.Fatalln("Could not get tree view selection:", err)
	}
	selection.SetMode(gtk.SELECTION_MULTIPLE)

	app.scannedListStore, err = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		log.Fatal("Unable to create list store:", err)
	}
	tv.SetModel(app.scannedListStore)

	var signals = map[string]interface{}{
		"menuOpen_activate_cb":                    app.menuOpen,
		"menuQuit_activate_cb":                    app.menuQuit,
		"inputBox_key_press_event_cb":             app.keyPress,
		"dlgNotFound_button_press_event_cb":       app.dlgNotFoundBtnPress,
		"dlgNotFound_key_press_event_cb":          app.dlgNotFoundKeyPress,
		"scannedBooksTreeView_key_press_event_cb": app.scannedBooksKeyPress,
		"btnShow_clicked_cb":                      app.btnShowClicked,
	}
	app.builder.ConnectSignals(signals)

	css, err := gtk.CssProviderNew()
	if err != nil {
		log.Fatal("Cannot create css provider:", err)
	}

	err = css.LoadFromData(cssData)
	if err != nil {
		log.Fatal("Cannot load css styles:", err)
	}

	app.addCSS(app.dlgAlreadySeen, css)
	app.addCSS(app.dlgExpiryDate, css)
	app.addCSS(app.dlgNotFound, css)

	wnd.AddEvents(int(gdk.KEY_PRESS_MASK))
	wnd.ShowAll()
	app.application.AddWindow(wnd)
}

func (app *Application) addCSS(w *gtk.Window, css *gtk.CssProvider) {
	style, err := w.GetStyleContext()
	if err != nil {
		log.Print("Cannot get style context:", err)
		return
	}
	style.AddProvider(css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func (app *Application) run() error {
	_, err := app.application.Connect("activate", app.builderFunc)
	if err != nil {
		return err
	}
	app.application.Run(os.Args)
	return nil
}

func NewApplication() *Application {
	gtkApp, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatalln("Couldn't create app:", err)
	}

	return &Application{
		application: gtkApp,
	}
}

func main() {

	app := NewApplication()

	err := app.run()
	if err != nil {
		log.Println("Error running application:", err)
	}
	// It looks like all builder code must execute in the context of `app`.
	// If you try creating the builder inside the main function instead of
	// the `app` "activate" callback, then you will get a segfault
}
