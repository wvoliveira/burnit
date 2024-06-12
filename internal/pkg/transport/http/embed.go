package http

import (
	"log"

	e "github.com/wvoliveira/burnit/internal/pkg/error"
	"github.com/wvoliveira/burnit/web"
)

func iconFile() ([]byte, error) {
	data, err := web.Embed.ReadFile("icon.png")
	if err != nil {
		log.Println("Error to read icon file:", err.Error())
		return nil, e.ErrInternalServer
	}
	return data, nil
}

func scriptFile() ([]byte, error) {
	data, err := web.Embed.ReadFile("script.js")
	if err != nil {
		log.Println("Error to read script file:", err.Error())
		return nil, e.ErrInternalServer
	}
	return data, nil
}

func indexFile() ([]byte, error) {
	data, err := web.Embed.ReadFile("index.html")
	if err != nil {
		log.Println("Error to read index file:", err.Error())
		return nil, e.ErrInternalServer
	}
	return data, nil
}
