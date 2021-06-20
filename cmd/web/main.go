package main

import (
	"context"
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
	ctx           context.Context //for stopping the keyer/tutor
	cancel        context.CancelFunc
	ktrunning     bool
	td            *templateData
	stationModel  *stationModel
}

func main() {
	var err error

	pw := flag.String("pw", "", "MySQL Password")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.LUTC)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Llongfile)

	//note, this requires the run command be issues from the project base
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	td := &templateData{}

	dsn := "web:" + *pw + "@/stationmaster?parseTime=true"
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
		td:            td,
		stationModel:  &stationModel{DB: db},
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
