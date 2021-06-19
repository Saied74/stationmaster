package main

import (
	"context"
	"database/sql"
	"flag"
	//	"fmt"
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

	td := &templateData{
		Top: LogsRow{"Time", "Call", "Mode", "Sent", "Rcvd", "Band", "Name", "Country",
			"Comment", "LOTW Sent", "LOTW Rcvd"},
		Table: []LogsRow{},
	}

	td.Table = append(td.Table, LogsRow{
		"3/30/2017", "5T2AI", "USB", "599", "599", "12m", "Al Graham", "Mauritania", "", "", "",
	})

	td.Table = append(td.Table, LogsRow{
		"3/23/2017", "J5UAP", "CW", "588", "599", "80m", "Peter Brucker", "Tunisia", "", "", "",
	})

	td.Table = append(td.Table, LogsRow{
		"1/13/17", "SV9BAT", "CW", "499", "579", "80m", "Ginnias Glanakis", "Crete", "", "", "",
	})

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
