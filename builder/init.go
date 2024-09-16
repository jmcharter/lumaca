package builder

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/jmcharter/lumaca/config"
)

func Initialise(author string, title string) error {
	// Check for existence of config.toml
	if _, err := os.Stat("config.toml"); err == nil {
		fmt.Println("Project already initialised")
		return nil // Early return if file exists
	}
	// Create config.toml
	if author == "" {
		author = "Blog Author"
	}
	if title == "" {
		title = "Blog Title"
	}
	var cfg_data config.Config
	cfg_data.Author.Name = author
	cfg_data.Site.Title = title
	cfg_data.Site.Author = author
	cfg_data.Directories.Posts = "content/posts"
	cfg_data.Directories.Pages = "content/pages"
	cfg_data.Directories.Static = "static"
	cfg_data.Directories.Templates = "templates"
	cfg_data.Directories.Dist = "dist"
	cfg_data.Files.Extension = ".html"

	f, err := os.Create("config.toml")
	if err != nil {
		return fmt.Errorf("failed to create config.toml: %w", err)
	}
	err = toml.NewEncoder(f).Encode(cfg_data)
	if err != nil {
		return fmt.Errorf("failed to write config data to config.toml: %w", err)
	}

	// Create "templates" and "content/static" directories with files
	err = os.MkdirAll("templates", 0755)
	if err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	err = os.MkdirAll("content/static", 0755)
	if err != nil {
		return fmt.Errorf("failed to create content/static directory: %w", err)
	}

	// Copy embedded files
	err = copyEmbeddedFiles(EmbeddedFiles, "templates", "templates")
	if err != nil {
		return err
	}
	err = copyEmbeddedFiles(EmbeddedFiles, "content/static", "content/static")
	if err != nil {
		return err
	}

	return nil
}

func copyEmbeddedFiles(fsys embed.FS, sourceDir, targetDir string) error {
	err := fs.WalkDir(fsys, sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relativePath)

		err = os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}

		err = os.WriteFile(targetPath, content, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		fmt.Printf("Copied %s to %s\n", path, targetPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy embedded files: %w", err)
	}
	return nil
}
