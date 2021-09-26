package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
)

const (
	buildDir = "docs"
)

//go:embed templates/chapter.html
var chapterTemplate string

//go:embed static/styles.css
var styles string

var tmpl = template.Must(template.New("chapter").Parse(chapterTemplate))

type Config struct {
	ID          string `json:"id"`
	ChapterGlob string `json:"chapterGlob"`
	location    string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	sources, err := filepath.Glob("./*/config.json")
	if err != nil {
		return err
	}

	for _, cfgFile := range sources {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return err
		}

		config := &Config{}
		if err := json.Unmarshal(b, config); err != nil {
			return err
		}
		config.location = cfgFile

		if err := buildBook(config); err != nil {
			return err
		}
	}

	return nil
}

func buildBook(config *Config) error {
	chapters, err := filepath.Glob(filepath.Join(filepath.Dir(config.location), config.ChapterGlob))
	if err != nil {
		return err
	}

	for _, chapter := range chapters {
		if err := buildChapter(config, chapter); err != nil {
			return err
		}
	}

	return nil
}

func buildChapter(config *Config, chapter string) error {
	contents, err := ioutil.ReadFile(chapter)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(contents, &buf); err != nil {
		return err
	}

	filename := filepath.Join(buildDir, chapter+".html")

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	variables := map[string]interface{}{
		"Title":   "Great Expectations",
		"Chapter": template.HTML(buf.String()),
		"Styles":  template.CSS(styles),
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, variables); err != nil {
		return err
	}

	return nil
}
