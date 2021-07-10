package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

//The design of this program is along the lines of Alex Edward's
//Let's Go except since it is a single user local program, it
//ignore the rules for a shared over the internet application

//for injecting data into handlers
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
	displayLines  int
	stationModel  *stationModel
	putCancel     putCancelFunc
	getCancel     getCancelFunc
	putId         putIdFunc
	getId         getIdFunc
	sKey          sessionMgr
	qrzuser       string
	qrzpw         string
}

func main() {
	var err error

	sqlpw := flag.String("sqlpw", "", "MySQL Password")
	displayLines := flag.Int("lines", 20, "No. of lines to be displayed on logs")
	qrzpw := flag.String("qrzpw", "", "QRZ.com Password")
	qrzuser := flag.String("qrzuser", "", "QRZ.com User Name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.LUTC)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Llongfile)

	//note, this requires the run command be issues from the project base
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	dsn := "web:" + *sqlpw + "@/stationmaster?parseTime=true"
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	putCancel, getCancel := contextStore()
	putId, getId := saveId()

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
		displayLines:  *displayLines,
		stationModel:  &stationModel{DB: db},
		putCancel:     putCancel,
		getCancel:     getCancel,
		putId:         putId,
		getId:         getId,
		sKey:          sessionCache(),
		qrzpw:         *qrzpw,
		qrzuser:       *qrzuser,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/home", app.home)
	mux.HandleFunc("/ktutor", app.ktutor)
	mux.HandleFunc("/qsolog", app.qsolog)
	mux.HandleFunc("/addlog", app.addlog)
	mux.HandleFunc("/ant", app.ant)
	mux.HandleFunc("/start", app.start)
	mux.HandleFunc("/stop", app.stop)
	mux.HandleFunc("/stopcode", app.stopcode)
	mux.HandleFunc("/editlog", app.editlog)
	mux.HandleFunc("/updatedb", app.updatedb)
	mux.HandleFunc("/quit", app.quit)
	mux.HandleFunc("/getconn", app.getConn)
	mux.HandleFunc("/callsearch", app.callSearch)
	mux.HandleFunc("/updateQRZ", app.updateQRZ)

	srv := &http.Server{
		Addr:     ":4000",
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("starting server on :4000")
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
