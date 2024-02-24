package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/frontmatter"
)

type Config struct {
	ContentDirectories struct {
		Posts string
		Pages string
	}
	Author struct {
		Name string
	}
}

type Matter struct {
	Title  string   `yaml:"title"`
	Author string   `yaml:"author"`
	Tags   []string `yaml:"tags"`
	Date   string
}

func main() {
	fmt.Println("Lumaca starting...")
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatal(err)
	}
	err := run(config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Lumaca finished.")
}

func run(config Config) error {

	files, err := os.ReadDir(config.ContentDirectories.Posts)
	if err != nil {
		return err
	}

	var matters []Matter
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			break
		}
		filepath := filepath.Join(config.ContentDirectories.Posts, file.Name())
		content, err := os.ReadFile(filepath)
		if err != nil {
			log.Println("Error reading file:", err)
			continue
		}

		var matter Matter
		_, err = frontmatter.Parse(strings.NewReader(string(content)), &matter)
		if err != nil {
			log.Println("Error parsing frontmatter from file:", err)
			continue
		}
		matters = append(matters, matter)

	}

	for _, m := range matters {
		fmt.Printf("%v\n", m)
	}

	return nil

}
