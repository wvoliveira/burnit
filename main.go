package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed icon.png index.html.tmpl
var files embed.FS

func main() {
	t, err := template.ParseFS(files, "index.html.tmpl")
	if err != nil {
		log.Fatal(err.Error())
	}

	data := struct {
		Title string
	}{
		Title: "Onetime",
	}

	mux := http.NewServeMux()
	mux.Handle("/", func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Println(fmt.Sprintf(
				`time=%s path=%s method=%s`,
				time.Now().Format(time.RFC3339), r.URL.Path, r.Method,
			))

			if r.URL.Path == "/icon.png" && r.Method == "GET" {
				file, _ := files.ReadFile("icon.png")
				_, _ = w.Write(file)
				return
			}

			if r.URL.Path == "/" && r.Method == "GET" {
				err := t.Execute(w, data)
				if err != nil {
					log.Println(fmt.Sprintf("Error: %s", err.Error()))
				}
				return
			}

			if r.URL.Path == "/" && r.Method == "POST" {
				// TODO
				return
			}

			w.WriteHeader(http.StatusNotFound)
			return
		})
	}(),
	)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	closeIdleConnections := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(closeIdleConnections)
	}()

	log.Println("HTTP server http://127.0.0.1:8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-closeIdleConnections
}
