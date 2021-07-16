package main

import (
	"io"
	"log"
)

func newTestApp() *application {

	templateCache, err := newTemplateCache("./../../ui/html/")
	if err != nil {
		log.Fatal(err)
	}
	return &application{
		infoLog:  log.New(io.Discard, "", 0),
		errorLog: log.New(io.Discard, "", 0),
		// errorLog:      log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Llongfile), //
		templateCache: templateCache,
	}
}

// type bodyType string
//
// func (b bodyType) Read(p []byte) (int, error) {
// 	p = []byte(b)
// 	return len(p), nil
// }
//
// func (b bodyType) Close() error {
// 	return nil
// }
