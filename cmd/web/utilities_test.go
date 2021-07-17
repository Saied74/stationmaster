package main

import (
	"io"
	"log"
	"net/http"
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
		logsModel:     &mockLogsModel{lastLogsErr: nil, defaultErr: nil},
		qrzModel:      &mockQRZModel{},
		otherModel:    &mockOtherModel{},
	}
}

func newMockClient(statcode int, r io.ReadCloser) httpClient {
	return &mockClient{
		mockGet: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}
}
