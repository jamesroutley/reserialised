package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/yuin/goldmark"
)

const (
	buildDir = "docs"
)

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

	filename := filepath.Join(buildDir, config.ID, chapter+".html")
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
