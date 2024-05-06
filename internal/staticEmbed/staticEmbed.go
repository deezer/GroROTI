package staticEmbed

import (
	"embed"
	"html/template"
	"io/fs"

	"github.com/rs/zerolog/log"
)

var (
	//go:embed templates/*
	EmbeddedTemplates embed.FS
	//go:embed static/*
	EmbeddedStatic embed.FS
	Templates      map[string]*template.Template
)

func LoadTemplates() error {
	if Templates == nil {
		Templates = make(map[string]*template.Template)
	}

	tmplFiles, err := fs.ReadDir(EmbeddedTemplates, "templates")
	if err != nil {
		log.Printf("readdir failed")
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		tmplPath := "templates/" + tmpl.Name()

		pt, err := template.ParseFS(EmbeddedTemplates, tmplPath)
		if err != nil {
			return err
		}

		Templates[tmplPath] = pt
	}
	return nil
}
