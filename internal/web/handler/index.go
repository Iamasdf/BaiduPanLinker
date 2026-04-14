package handler

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed index.html
var indexHTML embed.FS

func Index(c *gin.Context) {
	content, err := indexHTML.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load index.html")
		return
	}

	tmpl, err := template.New("index").Parse(string(content))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse template")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(c.Writer, nil)
}
