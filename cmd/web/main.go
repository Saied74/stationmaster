package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-yaml/yaml"
	"gobot.io/x/gobot/platforms/raspi"

	"github.com/Saied74/stationmaster/pkg/bandselect"
	"github.com/Saied74/stationmaster/pkg/vfo"
)

//The design of this program is along the lines of Alex Edward's
//Let's Go except since it is a single user local program, it
//ignore the rules for a shared over the internet application

type configType struct {
	DSN        string `yaml:"dsn"`
	ConfigFile string `yaml:"configfile"`
	ADIFFile   string `yaml:"adiffile"`
	QSLdir     string `yaml:"qsldir"`
	ContestDir string `yaml:"contestdir"`
}

//for injecting data into handlers
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
	displayLines  int
	logsModel     logsType
	qrzModel      qrzType
	otherModel    otherType
	putCancel     putCancelFunc
	getCancel     getCancelFunc
	putId         putIdFunc
	getId         getIdFunc
	sKey          sessionMgr
	qrzuser       string
	qrzpw         string
	adifFile      string
	qslDir        string
	contestDir    string
	vfoAdaptor    *raspi.Adaptor
	bandData      *bandselect.BandData
	cqStat		  [wsjtBuffer]int
	qsoStat		  [wsjtBuffer]int
	wsjtPntr	  int
}

type httpClient interface {
	Get(url string) (*http.Response, error)
}

type createFunction interface{}

var client httpClient

func init() {
	client = &http.Client{}
	writeControl = &fileWrite{}
	readControl = &fileRead{}
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

	var config = &configType{"", "", "", "", ""}
	configPath := os.Getenv("STATIONMASTER")
	configData, err := os.ReadFile(fmt.Sprintf("%s/config.yaml", configPath))
	if err != nil {
		errorLog.Fatal(err)
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		errorLog.Fatal(err)
	}

	//note, this requires the run command be issues from the project base
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	dsn := fmt.Sprintf(config.DSN, *sqlpw) //"web:" + *sqlpw + "@/stationmaster?parseTime=true"

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	putCancel, getCancel := contextStore()
	putId, getId := saveId()
	m := &otherModel{DB: db}

	home := os.Getenv("HOME")
	qslDir := strings.TrimPrefix(config.QSLdir, "$HOME/")
	contestDir := strings.TrimPrefix(config.ContestDir, "$HOME/")
	qslDir = filepath.Join(home, qslDir)
	contestDir = filepath.Join(home, contestDir)

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
		displayLines:  *displayLines,
		logsModel:     &logsModel{DB: db},
		qrzModel:      &qrzModel{DB: db},
		otherModel:    m,
		putCancel:     putCancel,
		getCancel:     getCancel,
		putId:         putId,
		getId:         getId,
		sKey:          m.sKey, //sessionCache(),
		qrzpw:         *qrzpw,
		qrzuser:       *qrzuser,
		adifFile:      fmt.Sprintf("%s/%s", qslDir, config.ADIFFile),
		qslDir:        qslDir,
		contestDir:    contestDir,
		vfoAdaptor:    vfo.Initvfo(171798692),
		cqStat:        [wsjtBuffer]int{},
		qsoStat:       [wsjtBuffer]int{},
		wsjtPntr:      0,

		//		bandData:      make(chan int), //bandselect.BandData.Band),
	}
	app.bandData = &bandselect.BandData{
		Band:    make(chan int),
		Adaptor: app.vfoAdaptor,
	}
	go bandselect.BandRead(app.bandData)
	go app.wsjtxServe()

	mux := app.routes()
	srv := &http.Server{
		Addr:     ":4000",
		ErrorLog: errorLog,
		Handler:  mux,
	}
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	infoLog.Printf("starting server on :4000")
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func (app *application) routes() *http.ServeMux {
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
	mux.HandleFunc("/defaults", app.defaults)
	mux.HandleFunc("/store-defaults", app.storeDefaults)
	mux.HandleFunc("/contacts", app.contacts)
	mux.HandleFunc("/adif", app.adif)
	mux.HandleFunc("/gen-adif", app.genadif)
	mux.HandleFunc("/cabrillo", app.cabrillo)
	mux.HandleFunc("/gencabrillo", app.genCabrillo)
	mux.HandleFunc("/analysis", app.analysis)
	mux.HandleFunc("/country", app.country)
	mux.HandleFunc("/country-confirmed", app.countryConfirmed)
	mux.HandleFunc("/countryselect", app.countrySelect)
	mux.HandleFunc("/county", app.county)
	mux.HandleFunc("/county-confirmed", app.countyConfirmed)
	mux.HandleFunc("/countyselect", app.countySelect)
	mux.HandleFunc("/repeat", app.repeat)
	mux.HandleFunc("/confirmqsls", app.confirmQSLs)
	mux.HandleFunc("/contacts-confirmed", app.contactsConfirmed)
	mux.HandleFunc("/state", app.state)
	mux.HandleFunc("/state-confirmed", app.stateConfirmed)
	mux.HandleFunc("/stateselect", app.stateSelect)
	mux.HandleFunc("/start-vfo", app.startVFO)
	mux.HandleFunc("/update-vfo", app.updateVFO)
	mux.HandleFunc("/update-band", app.updateBand)
	mux.HandleFunc("/update-dx", app.updateDX)
	return mux
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
