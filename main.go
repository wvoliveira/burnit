package main

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"io"
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
		Title: "BurnIt",
	}

	type responseBody struct {
		Status  string `json:"status"`
		Data    any    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	mux := http.NewServeMux()
	mux.Handle("/", func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf(
				`time=%s path=%s method=%s\n`,
				time.Now().Format(time.RFC3339), r.URL.Path, r.Method,
			)

			if (r.URL.Path == "/icon.png" || r.URL.Path == "/favicon.ico") && r.Method == "GET" {
				file, _ := files.ReadFile("icon.png")
				_, _ = w.Write(file)
				return
			}

			if r.URL.Path == "/" && r.Method == "GET" {
				err := t.Execute(w, data)
				if err != nil {
					log.Printf("Error: %s", err.Error())
				}
				return
			}

			if r.URL.Path == "/" && r.Method == "POST" {
				log.Println("Body", r.Body)

				maxMem := 10
				bytes := make([]byte, maxMem)
				content := []byte{}
				var size int

				for {
					read, err := r.Body.Read(bytes)
					size += read
					content = append(content, bytes...)

					if size >= 1024 {
						rb := responseBody{
							Status:  "error",
							Message: "Max size: 1Mb",
						}

						b, _ := json.Marshal(rb)
						w.Write(b)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					if err == io.EOF {
						log.Println("EOF!")
						break
					}
				}

				log.Printf("Size is %d\n", size)
				log.Printf("Bytes: %s\n", bytes)
				log.Printf("Content: %s\n", content)

				// OK
				// defer r.Body.Close()
				// f, e := os.Create("file.txt")
				// if e != nil {
				// 	panic(e)
				// }
				// defer f.Close()
				// f.ReadFrom(r.Body)

				// copy example
				// f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
				// if err != nil {
				// 	panic(err) //please dont
				// }
				// defer f.Close()
				// io.Copy(f, file)

				return
			}

			w.WriteHeader(http.StatusNotFound)
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
