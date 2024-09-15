package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/jmcharter/lumaca/config"
)

func New(cfg config.Config, title string, author string, draft bool) error {
	if author == "" {
		author = cfg.Author.Name
	}
	metadata := Matter{
		Title:   title,
		Author:  author,
		Tags:    []string{},
		Date:    YAMLDate(time.Now()),
		Type:    contentTypePost,
		Slug:    newPostSlug(title),
		IsDraft: draft,
	}
	contentDir := cfg.Directories.Posts
	filename := fmt.Sprintf("%s.md", metadata.Slug.String())
	filepath := filepath.Join(contentDir, filename)

	frontmatterBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter to YAML: %w", err)
	}
	frontmatter := fmt.Sprintf("---\n%s---\n\n", string(frontmatterBytes))

	initialContent := frontmatter + "\n#" + title + "\n\nLorem ipsum..."
	err = os.WriteFile(filepath, []byte(initialContent), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create content file: %w", err)
	}
	fmt.Printf("New file created: %s", filepath)
	return nil
}

func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	for i, tag := range tags {
		tags[i] = fmt.Sprintf("\"%s\"", tag)
	}
	return strings.Join(tags, ", ")
}
