package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	kitendoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/wvoliveira/burnit/internal/pkg/endpoint"
	e "github.com/wvoliveira/burnit/internal/pkg/error"
)

func NewHTTPHandler(endpoints endpoint.Set, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	m := http.NewServeMux()

	// HTTP endpoints for embeded web app.
	// Open app in your browser to look the "nice" web interface.
	// m.HandleFunc("/", keyHandler).Methods("GET").Queries("key", "{key}")
	// m.HandleFunc("/", rootHandler)
	m.HandleFunc("/icon.png", iconFileHandler)
	m.HandleFunc("/favicon.ico", iconFileHandler)
	m.HandleFunc("/script.js", scriptFileHandler)

	// HTTP API REST endpoints.
	// The entry point of the API becomes here.
	// m.HandleFunc("/api/content", createContentHandler).Methods("POST")

	// HTTP endpoints for app healthcheck.
	// Useful for most of orchestration tools.
	m.HandleFunc("/api/info", infoHandler)
	m.HandleFunc("/api/healthcheck", healthcheckHandler)
	m.HandleFunc("/api/healthcheck/live", healthcheckHandler)
	m.HandleFunc("/api/healthcheck/ready", healthcheckHandler)

	// Handlers to help to test some HTTP functions. Ex.: Graceful shutdown
	m.HandleFunc("/api/test/delay", delayHandler)

	m.Handle("/api/content", httptransport.NewServer(
		endpoints.CreateEndpoint,
		decodeHTTPCreateRequest,
		encodeHTTPGenericResponse,
		options...,
	))

	return m

}

// decodeHTTPCreateRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded sum request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.CreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(kitendoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println("Error:", err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

func iconFileHandler(w http.ResponseWriter, r *http.Request) {
	file, err := iconFile()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			errorHandler(w, r, err)
		}
	}

	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			errorHandler(w, r, err)
		}
	}
}

func scriptFileHandler(w http.ResponseWriter, r *http.Request) {
	file, err := scriptFile()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			errorHandler(w, r, err)
		}
	}

	w.Header().Add("Content-type", "text/javascript;chartset=utf-8")
	_, err = w.Write(file)
	if err != nil {
		fmt.Println("Error to write:", err.Error())
		return
	}
}

// func keyHandler(w http.ResponseWriter, r *http.Request) {
// 	key := r.URL.Query().Get("key")
// 	fmt.Println("URL query key: ", key)

// 	c, err := keyContent(key)
// 	if err != nil {
// 		w.WriteHeader(404)

// 		_, err := w.Write([]byte(err.Error()))
// 		if err != nil {
// 			fmt.Println("Error to write:", err.Error())
// 			return
// 		}
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(200)

// 	rb := struct {
// 		Status   string         `json:"status"`
// 		Key      string         `json:"key"`
// 		Text     []byte         `json:"text"`
// 		FileName string         `json:"filename"`
// 		File     multipart.File `json:"-"`
// 	}{
// 		Status:   "ok",
// 		Key:      c.Key,
// 		Text:     c.Text,
// 		FileName: c.FileName,
// 		File:     c.File,
// 	}
// 	responseBytes, err := json.Marshal(rb)
// 	if err != nil {
// 		errorHandler(w, r, err)
// 		return
// 	}

// 	_, err = w.Write(responseBytes)
// 	if err != nil {
// 		fmt.Println("Error to write:", err.Error())
// 		return
// 	}
// }

// func rootHandler(w http.ResponseWriter, r *http.Request) {
// 	q := r.URL.Query()
// 	if q.Get("key") != "" {
// 		keyHandler(w, r)
// 		return
// 	}

// 	data, err := indexFile()
// 	if err != nil {
// 		fmt.Println("Error:", err.Error())
// 		w.WriteHeader(500)

// 		_, err := w.Write([]byte(err.Error()))
// 		if err != nil {
// 			fmt.Println("Error to write:", err.Error())
// 			return
// 		}
// 		return
// 	}

// 	_, err = w.Write(data)
// 	if err != nil {
// 		fmt.Println("Error:", err.Error())
// 		w.WriteHeader(500)

// 		_, err := w.Write([]byte(err.Error()))
// 		if err != nil {
// 			fmt.Println("Error to write:", err.Error())
// 			return
// 		}
// 		return
// 	}
// 	return
// }

// func createContentHandler(w http.ResponseWriter, r *http.Request) {
// 	err := r.ParseMultipartForm(1 << 20) // 1 MB limit
// 	if err != nil {
// 		http.Error(w, "Unable to parse form", http.StatusBadRequest)
// 		return
// 	}

// 	file, fileHandler, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "Unable to get the file", http.StatusBadRequest)
// 		return
// 	}
// 	defer func(file multipart.File) {
// 		_ = file.Close()
// 	}(file)

// 	// Get the content from the textarea
// 	textContent := r.FormValue("text")
// 	fmt.Printf("Received content: %s\n", textContent)

// 	size := len(textContent)
// 	fmt.Println("Length:", size)

// 	if size > 1000 {
// 		rb := struct {
// 			Status  string `json:"status"`
// 			Message string `json:"message"`
// 		}{
// 			Status:  "error",
// 			Message: "Max size: 1000 bytes",
// 		}

// 		b, err := json.Marshal(rb)
// 		if err != nil {
// 			errorHandler(w, r, err)
// 		}

// 		w.WriteHeader(http.StatusBadRequest)

// 		_, err = w.Write(b)
// 		if err != nil {
// 			errorHandler(w, r, err)
// 		}
// 		return
// 	}

// 	fmt.Printf("Size is %d\n", size)
// 	fmt.Printf("Content: %s\n", textContent)

// 	// key, err := createKey(fileHandler.Filename, file, []byte(textContent))
// 	if err != nil {
// 		errorHandler(w, r, err)
// 	}

// 	rb := struct {
// 		Status  string `json:"status"`
// 		Message string `json:"message"`
// 	}{
// 		Status:  "ok",
// 		Message: "key",
// 	}

// 	b, err := json.Marshal(rb)
// 	if err != nil {
// 		errorHandler(w, r, err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	_, err = w.Write(b)
// 	if err != nil {
// 		errorHandler(w, r, err)
// 	}
// 	return
// }

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	rb := struct {
		AppVersion string `json:"app_version"`
		GoVersion  string `json:"go_version"`
	}{
		AppVersion: "ok",
		GoVersion:  runtime.Version(),
	}

	content, err := json.Marshal(rb)
	if err != nil {
		errorHandler(w, r, err)
	}

	_, _ = w.Write(content)
}

func delayHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	switch err {
	case e.ErrBadRequest:
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

type errorWrapper struct {
	Error string `json:"error"`
}
