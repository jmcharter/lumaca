package builder

import (
	"errors"
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

	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gosimple/slug"
	"github.com/jmcharter/lumaca/config"
)

type YAMLDate time.Time

func (d YAMLDate) MarshalYAML() (interface{}, error) {
	return time.Time(d).Format("2006-01-02"), nil
}

func (d *YAMLDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dateStr string
	if err := unmarshal(&dateStr); err != nil {
		return err
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}
	*d = YAMLDate(parsedDate)
	return nil
}

type postSlug string

func newPostSlug(s string) postSlug {
	return postSlug(slug.Make(s))
}

func (s postSlug) String() string {
	return string(s)
}

func getTemplateFilePath(config config.Config, cType contentType) string {
	var templateName = cType.String()
	return filepath.Join(config.Directories.Templates, templateName+config.Files.Extension)
}

func getOutputFilePath(config config.Config, name string, cType contentType) string {
	var baseDir string
	switch cType {
	case contentTypeBase:
		log.Fatal("Base content has no output file path")
	case contentTypeHome:
		baseDir = config.Directories.Dist
	case contentTypePage:
		baseDir = config.Directories.Pages
	case contentTypePost:
		baseDir = config.Directories.Posts
	}
	outputDirPath := filepath.Join(config.Directories.Dist, filepath.Base(baseDir))
	return filepath.Join(outputDirPath, name+config.Files.Extension)
}

func getPageOutputFilePath(config config.Config, title string) string {
	outputDirPath := filepath.Join(config.Directories.Dist, filepath.Base(config.Directories.Pages))
	return filepath.Join(outputDirPath, title+config.Files.Extension)
}

func getIndexPath(config config.Config) string {
	return filepath.Join(config.Directories.Dist, "index"+config.Files.Extension)
}

type contentType int32

const (
	contentTypeBase contentType = iota
	contentTypeHome
	contentTypePage
	contentTypePost
)

var contentTypeTemplates = [...]string{
	"base",
	"home",
	"page",
	"post",
}

func (c contentType) String() string {
	if c < contentTypeBase || c > contentTypePost {
		log.Fatal("Content Type not implemented")
	}
	return contentTypeTemplates[c]
}

func (c contentType) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *contentType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	for i, v := range contentTypeTemplates {
		if v == s {
			*c = contentType(i)
			return nil
		}
	}
	return errors.New("invalid content type")
}

type Matter struct {
	Title   string      `yaml:"title"`
	Author  string      `yaml:"author"`
	Tags    []string    `yaml:"tags"`
	Date    YAMLDate    `yaml:"date"`
	Type    contentType `yaml:"type"`
	Slug    postSlug    `yaml:"-"`
	IsDraft bool        `yaml:"draft"`
}

type MarkdownData struct {
	Frontmatter Matter
	Content     []byte
	HTMLContent template.HTML
	Path        string
}

type SiteData struct {
	Title  string
	Author string
	Pages  []MarkdownData
}

func Build(config config.Config) {
	fmt.Println("Build starting...")
	run(config)
	fmt.Println("Build finished.")
}

func run(config config.Config) {
	err := makeDirs(config)
	if err != nil {
		log.Fatal("failed to make directories:", err)
	}
	postMDData, err := getMarkdownData(config, contentTypePost, config.Directories.Posts)
	if err != nil {
		log.Fatal(err)
	}
	postMarkdown := RenderAllMDToHTML(postMDData)
	pageMDData, err := getMarkdownData(config, contentTypePage, config.Directories.Pages)
	pageMarkdown := RenderAllMDToHTML(pageMDData)
	siteData := SiteData{
		Title:  config.Site.Title,
		Author: config.Site.Author,
	}
	err = renderPages(config, pageMarkdown, &siteData)
	if err != nil {
		log.Fatal(err)
	}
	err = renderPosts(config, postMarkdown, &siteData)
	if err != nil {
		log.Fatal(err)
	}
	err = renderHome(config, postMarkdown, &siteData)
	if err != nil {
		log.Fatal(err)
	}
	err = copyStaticDir(config)
	if err != nil {
		log.Fatal(err)
	}

}

func makeDirs(config config.Config) error {
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

func renderPages(config config.Config, mds []MarkdownData, siteData *SiteData) error {
	err := executeTemplates(config, mds, siteData, contentTypePage)
	return err
}

func renderPosts(config config.Config, mds []MarkdownData, siteData *SiteData) error {
	if len(mds) < 1 {
		return fmt.Errorf("Site data contained no Markdown data")
	}
	err := executeTemplates(config, mds, siteData, contentTypePost)
	return err
}

func renderHome(config config.Config, mds []MarkdownData, siteData *SiteData) error {
	// Home template will inherit from base
	baseTmplFilePath := getTemplateFilePath(config, contentTypeBase)
	homeTmplFilePath := getTemplateFilePath(config, contentTypeHome)
	outputFilePath := getIndexPath(config)
	tmpl, err := template.ParseFiles(baseTmplFilePath, homeTmplFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse files: %w", err)
	}
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	data := struct {
		MD       []MarkdownData
		SiteData *SiteData
	}{
		mds,
		siteData,
	}
	err = tmpl.Execute(outputFile, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}

func executeTemplates(config config.Config, mds []MarkdownData, siteData *SiteData, cType contentType) error {
	for i, md := range mds {
		// Post template will inherit from base template
		baseTmplFilePath := getTemplateFilePath(config, contentTypeBase)
		contentTmplFilePath := getTemplateFilePath(config, cType)
		tmpl, err := template.ParseFiles(baseTmplFilePath, contentTmplFilePath)
		if err != nil {
			return fmt.Errorf("failed to parse templates: %w", err)
		}
		var fileName string

		switch cType {
		case contentTypePost:
			fileName = md.Frontmatter.Slug.String()
		default:
			fileName = md.Frontmatter.Title
		}

		outputFilePath := getOutputFilePath(config, fileName, cType)
		mds[i].Path, err = filepath.Rel(config.Directories.Dist, outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to generate relative path for output file: %w", err)
		}
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		data := struct {
			MD       MarkdownData
			SiteData *SiteData
		}{
			md,
			siteData,
		}
		err = tmpl.Execute(outputFile, data)
		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
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

func copyStaticDir(config config.Config) error {
	staticDir := config.Directories.Static
	destDir := filepath.Join(config.Directories.Dist, filepath.Base(staticDir))
	err := copyDir(staticDir, destDir)
	return err
}

// Iterates through the given directory and extracts Frontmatter and content from Markdown files
func getMarkdownData(config config.Config, cType contentType, inputDir string) ([]MarkdownData, error) {
	files, err := os.ReadDir(inputDir)
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
			filepath := filepath.Join(inputDir, file.Name())
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
			if cType == contentTypePost {
				if matter.Author == "" {
					matter.Author = config.Author.Name
				}
				matter.Slug = newPostSlug(matter.Title)
			}
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
