package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
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

	for i, chapter := range chapters {
		if err := buildChapter(i+1, config, chapter); err != nil {
			return err
		}
	}

	return nil
}

func buildChapter(number int, config *Config, chapter string) error {
	contents, err := ioutil.ReadFile(chapter)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(contents, &buf); err != nil {
		return err
	}

	filename := fmt.Sprintf("%02d-%s.html", number, chapterHash("Great Expectations", number))

	path := filepath.Join(buildDir, "great-expectations", filename)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	type previousChapter struct {
		Name string
		URL  string
	}

	var previousChapters []*previousChapter
	for i := 1; i < number; i++ {
		chapterName := fmt.Sprintf("%02d-%s.html", i, chapterHash("Great Expectations", i))
		previousChapters = append(previousChapters, &previousChapter{
			Name: fmt.Sprintf("Chapter %d", i),
			URL:  chapterName,
		})
	}

	variables := map[string]interface{}{
		"Title":            "Great Expectations",
		"Chapter":          template.HTML(buf.String()),
		"Styles":           template.CSS(styles),
		"PreviousChapters": previousChapters,
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, variables); err != nil {
		return err
	}

	return nil
}

func chapterHash(bookName string, chapterNumber int) string {
	return hash(fmt.Sprintf("reserialised-%s-%d", bookName, chapterNumber))
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}
