package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gosimple/slug"
)

type Config struct {
	Directories struct {
		Posts     string
		Pages     string
		Templates string
		Dist      string
	}
	Author struct {
		Name string
	}
	Files struct {
		Extension string
	}
	Site struct {
		Title string
	}
}

type ContentType string
type Slug string

func NewSlug(s string) Slug {
	return Slug(slug.Make(s))
}

func getTemplateFilePath(config Config, templateName string) string {
	return filepath.Join(config.Directories.Templates, templateName+config.Files.Extension)
}

func getPostOutputFilePath(config Config, slug Slug) string {
	outputDirPath := filepath.Join(config.Directories.Dist, filepath.Base(config.Directories.Posts))
	return filepath.Join(outputDirPath, string(slug)+config.Files.Extension)
}

func getIndexPath(config Config) string {
	return filepath.Join(config.Directories.Dist, "index"+config.Files.Extension)
}

const (
	ContentTypePost ContentType = "Post"
)

type Matter struct {
	Title  string    `yaml:"title"`
	Author string    `yaml:"author"`
	Tags   []string  `yaml:"tags"`
	Date   time.Time `yaml:"date"`
	Type   ContentType
	Slug   Slug
}

type MarkdownData struct {
	Frontmatter Matter
	Content     []byte
	HTMLContent template.HTML
	Path        string
}

func main() {
	fmt.Println("Lumaca starting...")
	config := initConfig()
	run(config)
	fmt.Println("Lumaca finished.")
}

func initConfig() Config {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatal(err)
	}
	return config
}

func run(config Config) {
	markdownData, err := getMarkdownData(config)
	if err != nil {
		log.Fatal(err)
	}
	convertedMarkdown := RenderAllMDToHTML(markdownData)
	markdownWithPaths := renderPosts(convertedMarkdown, config)
	renderHome(markdownWithPaths, config)

}

func renderPosts(convertedMarkdown []MarkdownData, config Config) []MarkdownData {
	for i, md := range convertedMarkdown {
		baseTmplFilepath := filepath.Join(config.Directories.Templates, "base.html")
		postTmplFilepath := filepath.Join(config.Directories.Templates, "post.html")
		tmpl, err := template.ParseFiles(baseTmplFilepath, postTmplFilepath)
		if err != nil {
			log.Fatal(err)
		}
		outputDirpath := filepath.Join(config.Directories.Dist, "posts")
		os.MkdirAll(outputDirpath, os.ModePerm)
		outputFilepath := filepath.Join(outputDirpath, string(md.Frontmatter.Slug)+".html")
		outputFile, err := os.Create(outputFilepath)
		if err != nil {
			log.Fatal(err)
		}
		convertedMarkdown[i].Path = filepath.Join("posts", string(md.Frontmatter.Slug)+".html")
		err = tmpl.Execute(outputFile, md)
		if err != nil {
			log.Fatal(err)
		}
	}
	return convertedMarkdown
}

func renderHome(posts []MarkdownData, config Config) {
	baseTmplFilepath := filepath.Join(config.Directories.Templates, "base.html")
	homeTmplFilepath := filepath.Join(config.Directories.Templates, "home.html")
	outputDirpath := filepath.Join(config.Directories.Dist)
	outputFilepath := filepath.Join(outputDirpath, "index.html")
	tmpl, err := template.ParseFiles(baseTmplFilepath, homeTmplFilepath)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(outputDirpath, os.ModePerm)

	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(outputFile, posts)
	if err != nil {
		log.Fatal(err)
	}
}

func getMarkdownData(config Config) ([]MarkdownData, error) {
	files, err := os.ReadDir(config.Directories.Posts)
	if err != nil {
		return nil, err
	}

	var contentFiles []MarkdownData
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		filepath := filepath.Join(config.Directories.Posts, file.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			log.Println("Error reading file:", err)
			continue
		}

		var matter Matter
		content, err := frontmatter.Parse(strings.NewReader(string(data)), &matter)
		if err != nil {
			log.Println("Error parsing frontmatter from file:", err)
			continue
		}
		if matter.Type == "" {
			matter.Type = ContentTypePost
		}
		matter.Slug = NewSlug(matter.Title)
		fileData := MarkdownData{
			Frontmatter: matter,
			Content:     content,
		}
		contentFiles = append(contentFiles, fileData)

	}

	return contentFiles, nil

}

func RenderAllMDToHTML(mds []MarkdownData) []MarkdownData {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.SuperSubscript

	htmlFlags := html.CommonFlags | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}

	for i := range mds {
		parser := parser.NewWithExtensions(extensions)
		doc := parser.Parse(mds[i].Content)
		renderer := html.NewRenderer(opts)
		mds[i].HTMLContent = template.HTML(markdown.Render(doc, renderer))
	}

	return mds
}
