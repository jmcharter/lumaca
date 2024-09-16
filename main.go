package main

import (
	"embed"

	"github.com/jmcharter/lumaca/builder"
	"github.com/jmcharter/lumaca/cmd"
)

//go:embed templates/* content/static/*
var embeddedFiles embed.FS

func main() {
	builder.EmbeddedFiles = embeddedFiles
	cmd.Execute()
}
