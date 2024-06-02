package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"runtime"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			`time=%s path=%s method=%s`,
			time.Now().Format(time.RFC3339), r.URL.Path, r.Method,
		)
		next.ServeHTTP(w, r)
	})
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

	_, err = w.Write(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		errorHandler(w, r, err)
	}
}

func scriptFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/javascript;chartset=utf-8")
	file, err := scriptFile()
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

	_, err = w.Write([]byte(content))
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
	log.Println("Body", r.Body)

	maxMem := 10
	bytes := make([]byte, maxMem)
	var content []byte
	var size int

	for {
		read, err := r.Body.Read(bytes)
		size += read
		content = append(content, bytes...)

		if size > 1000 {
			rb := responseBody{
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

		if err == io.EOF {
			break
		}
	}

	log.Printf("Size is %d\n", size)
	log.Printf("Content: %s\n", content)

	key, err := createKey(content)
	if err != nil {
		errorHandler(w, r, err)
	}

	rb := responseBody{
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
	body := responseInfoBody{
		AppVersion:    AppVersion,
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

func Router() *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.HandleFunc("/icon.png", iconFileHandler).Methods("GET")
	r.HandleFunc("/favicon.ico", iconFileHandler).Methods("GET")
	r.HandleFunc("/script.js", scriptFileHandler).Methods("GET")

	r.HandleFunc("/", keyHandler).Methods("GET").Queries("key", "{key}")
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/", createContentHandler).Methods("POST")

	r.HandleFunc("/info", infoHandler).Methods("GET")
	r.HandleFunc("/healthcheck", healthcheckHandler).Methods("GET")
	r.HandleFunc("/healthcheck/live", healthcheckHandler).Methods("GET")
	r.HandleFunc("/healthcheck/ready", healthcheckHandler).Methods("GET")

	// Handlers to help to test some theories. Ex.: graceful shutdown
	r.HandleFunc("/test/delay", delayHandler).Methods("GET")
	return r
}
