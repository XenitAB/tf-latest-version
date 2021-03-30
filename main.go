package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/pkg/helm"
	"github.com/xenitab/tf-provider-latest/pkg/provider"
	"github.com/xenitab/tf-provider-latest/pkg/update"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("expected path input")
		os.Exit(1)
	}
	path := args[0]

	// Disable Terraform logs
	log.SetOutput(ioutil.Discard)

	fs := afero.NewOsFs()
	providerResults, err := provider.Update(fs, string(path))
	if err != nil {
		fmt.Printf("failed updating provider versions: %q\n", err)
		os.Exit(1)
	}
	providerMd, err := update.ToMarkdown("Provider", providerResults)
	if err != nil {
		fmt.Printf("failed rendering provider markdown: %q\n", err)
		os.Exit(1)
	}
	helmResults, err := helm.Update(fs, string(path))
	if err != nil {
		fmt.Printf("failed updating helm versions: %q\n", err)
		os.Exit(1)
	}
	helmMd, err := update.ToMarkdown("Helm", helmResults)
	if err != nil {
		fmt.Printf("failed rendering provider markdown: %q\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n\n%s\n", providerMd, helmMd)
}
