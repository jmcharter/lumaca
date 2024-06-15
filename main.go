package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
		Static    string
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
		Title  string
		Author string
	}
}

type ContentType string
type Slug string

func NewSlug(s string) Slug {
	return Slug(slug.Make(s))
}

func (s Slug) String() string {
	return string(s)
}

func getTemplateFilePath(config Config, templateName string) string {
	return filepath.Join(config.Directories.Templates, templateName+config.Files.Extension)
}

func getPostOutputFilePath(config Config, slug Slug) string {
	outputDirPath := filepath.Join(config.Directories.Dist, filepath.Base(config.Directories.Posts))
	return filepath.Join(outputDirPath, slug.String()+config.Files.Extension)
}

func getIndexPath(config Config) string {
	return filepath.Join(config.Directories.Dist, "index"+config.Files.Extension)
}

const (
	ContentTypePost ContentType = "post"
)

type Matter struct {
	Title   string    `yaml:"title"`
	Author  string    `yaml:"author"`
	Tags    []string  `yaml:"tags"`
	Date    time.Time `yaml:"date"`
	Type    ContentType
	Slug    Slug
	IsDraft bool
}

type MarkdownData struct {
	Frontmatter Matter
	Content     []byte
	HTMLContent template.HTML
	Path        string
}

type SiteData struct {
	MD     []MarkdownData
	Title  string
	Author string
}

func main() {
	fmt.Println("Lumaca starting...")
	config, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	run(config)
	fmt.Println("Lumaca finished.")
}

func initConfig() (Config, error) {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}
	return config, nil
}

func run(config Config) {
	err := makeDirs(config)
	if err != nil {
		log.Fatal("failed to make directories:", err)
	}
	markdownData, err := getMarkdownData(config)
	if err != nil {
		log.Fatal(err)
	}
	convertedMarkdown := RenderAllMDToHTML(markdownData)
	siteData := SiteData{
		MD:     convertedMarkdown,
		Title:  config.Site.Title,
		Author: config.Site.Author,
	}
	err = renderPosts(&siteData, config)
	if err != nil {
		log.Fatal(err)
	}
	err = renderHome(&siteData, config)
	if err != nil {
		log.Fatal(err)
	}
	err = copyStaticDir(config)
	if err != nil {
		log.Fatal(err)
	}

}

func makeDirs(config Config) error {
	outputDirPath := config.Directories.Dist
	outputPostsPath := filepath.Join(outputDirPath, filepath.Base(config.Directories.Posts))
	outputStaticPath := filepath.Join(outputDirPath, "static")
	err := os.MkdirAll(outputPostsPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create posts directory")
	}
	err = os.MkdirAll(outputStaticPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create static directory")
	}
	filepath.WalkDir(config.Directories.Static, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error accessing path %q: %w", path, err)
		}
		if !d.IsDir() || path == config.Directories.Static {
			return nil
		}
		err = os.MkdirAll(filepath.Join(outputStaticPath, d.Name()), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create static subdirectory: %w", err)
		}
		return nil
	})
	return nil
}

func renderPosts(siteData *SiteData, config Config) error {
	for i, md := range siteData.MD {
		// Post template will inherit from base template
		baseTmplFilePath := getTemplateFilePath(config, "base")
		postTmplFilePath := getTemplateFilePath(config, "post")
		tmpl, err := template.ParseFiles(baseTmplFilePath, postTmplFilePath)
		if err != nil {
			return fmt.Errorf("failed to parse templates: %w", err)
		}
		outputFilePath := getPostOutputFilePath(config, md.Frontmatter.Slug)
		siteData.MD[i].Path, err = filepath.Rel(config.Directories.Dist, outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to generate relative path for output file: %w", err)
		}
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		err = tmpl.Execute(outputFile, md)
		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
	}
	return nil
}

func renderHome(siteData *SiteData, config Config) error {
	// Home template will inherit from base
	baseTmplFilePath := getTemplateFilePath(config, "base")
	homeTmplFilePath := getTemplateFilePath(config, "home")
	outputFilePath := getIndexPath(config)
	tmpl, err := template.ParseFiles(baseTmplFilePath, homeTmplFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse files: %w", err)
	}
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	err = tmpl.Execute(outputFile, siteData)
	if err != nil {
		return fmt.Errorf("failed to execture template: %w", err)
	}
	return nil
}

// make copydir func for recursion in copyStaticDir
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get directory info for src dir: %w", err)
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to make dst directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read src dir: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("failed to copy dir: %w", err)
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}

		}
	}

	return nil
}

func copyFile(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get file info for src file: %w", err)
	}

	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create dst file: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func copyStaticDir(config Config) error {
	staticDir := config.Directories.Static
	destDir := filepath.Join(config.Directories.Dist, filepath.Base(staticDir))
	err := copyDir(staticDir, destDir)
	return err
}

// Iterates through the Posts directory and extracts Frontmatter and content from Markdown files
func getMarkdownData(config Config) ([]MarkdownData, error) {
	files, err := os.ReadDir(config.Directories.Posts)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var contentFiles []MarkdownData

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		wg.Add(1)
		go func(file os.DirEntry) {
			defer wg.Done()
			filepath := filepath.Join(config.Directories.Posts, file.Name())
			data, err := os.ReadFile(filepath)
			if err != nil {
				log.Println("Error reading file:", err)
				return
			}

			var matter Matter
			content, err := frontmatter.Parse(strings.NewReader(string(data)), &matter)
			if err != nil {
				log.Println("Error parsing frontmatter from file:", err)
				return
			}
			if matter.Author == "" {
				matter.Author = config.Author.Name
			}
			if matter.Type == "" {
				matter.Type = ContentTypePost
			}
			matter.Slug = NewSlug(matter.Title)
			fileData := MarkdownData{
				Frontmatter: matter,
				Content:     content,
			}
			mu.Lock()
			contentFiles = append(contentFiles, fileData)
			mu.Unlock()

		}(file)
	}
	wg.Wait()

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
