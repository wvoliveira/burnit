package http

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/wvoliveira/burnit/internal/pkg/config"
	"github.com/wvoliveira/burnit/internal/pkg/endpoint"
	"github.com/wvoliveira/burnit/internal/pkg/model"
)

func NewHTTPHandler(endpoints endpoint.Set, logger *log.Logger) *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Use(headersMiddleware)

	// HTTP endpoints for embeded web app.
	// Open app in your browser to look the "nice" web interface.
	r.HandleFunc("/", keyHandler).Methods("GET").Queries("key", "{key}")
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/icon.png", iconFileHandler).Methods("GET")
	r.HandleFunc("/favicon.ico", iconFileHandler).Methods("GET")
	r.HandleFunc("/script.js", scriptFileHandler).Methods("GET")

	// HTTP API REST endpoints.
	// The entry point of the API becomes here.
	r.HandleFunc("/api/content", createContentHandler).Methods("POST")

	// HTTP endpoints for app healthcheck.
	// Useful for most of orchestration tools.
	r.HandleFunc("/api/info", infoHandler).Methods("GET")
	r.HandleFunc("/api/healthcheck", healthcheckHandler).Methods("GET")
	r.HandleFunc("/api/healthcheck/live", healthcheckHandler).Methods("GET")
	r.HandleFunc("/api/healthcheck/ready", healthcheckHandler).Methods("GET")

	// Handlers to help to test some HTTP functions. Ex.: Graceful shutdown
	r.HandleFunc("/api/test/delay", delayHandler).Methods("GET")
	return r
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("Error:", err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
	return
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
		log.Println("Error to write:", err.Error())
		return
	}
	return
}

func keyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Println("URL query key: ", key)

	content, err := keyContent(key)
	if err != nil {
		w.WriteHeader(404)

		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			log.Println("Error to write:", err.Error())
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	response := model.ResponseBodyJSON{
		Status:   "ok",
		Key:      content.Key,
		Text:     content.Text,
		FileName: content.FileName,
		File:     content.File,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		errorHandler(w, r, err)
		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		log.Println("Error to write:", err.Error())
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	data, err := indexFile()
	if err != nil {
		log.Printf("Error: %s", err.Error())
		w.WriteHeader(500)

		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			log.Println("Error to write:", err.Error())
			return
		}
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		w.WriteHeader(500)

		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			log.Println("Error to write:", err.Error())
			return
		}
		return
	}
	return
}

func createContentHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 20) // 1 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, fileHandler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to get the file", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	// Get the content from the textarea
	textContent := r.FormValue("text")
	fmt.Printf("Received content: %s\n", textContent)

	size := len(textContent)
	fmt.Println("Length:", size)

	if size > 1000 {
		rb := model.ResponseBody{
			Status:  "error",
			Message: "Max size: 1000 bytes",
		}

		b, err := json.Marshal(rb)
		if err != nil {
			errorHandler(w, r, err)
		}

		w.WriteHeader(http.StatusBadRequest)

		_, err = w.Write(b)
		if err != nil {
			errorHandler(w, r, err)
		}
		return
	}

	log.Printf("Size is %d\n", size)
	log.Printf("Content: %s\n", textContent)

	key, err := createKey(fileHandler.Filename, file, []byte(textContent))
	if err != nil {
		errorHandler(w, r, err)
	}

	rb := model.ResponseBody{
		Status:  "ok",
		Message: key,
	}

	b, err := json.Marshal(rb)
	if err != nil {
		errorHandler(w, r, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		errorHandler(w, r, err)
	}
	return
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	body := model.ResponseInfoBody{
		AppVersion:    config.AppVersion,
		GolangVersion: runtime.Version(),
	}

	content, err := json.Marshal(body)
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
