package main

import (
	"log"

	"github.com/nikhilsbhat/helm-images/cmd"
	"github.com/spf13/cobra/doc"
)

//go:generate go run github.com/nikhilsbhat/helm-images/docs
func main() {
	commands := cmd.SetImagesCommands()

	if err := doc.GenMarkdownTree(commands, "doc"); err != nil {
		log.Fatal(err)
	}
}
