package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"github.com/xenitab/tf-provider-latest/internal/update"
)

func main() {
	// Disable Terraform logs
	log.SetOutput(ioutil.Discard)

	// Parse flags
	path := flag.String("path", "", "path where directory recursion should start")
	providerSelector := flag.StringSlice("provider-selector", nil, "optional selector for providers to update")
	helmSelector := flag.StringSlice("helm-selector", nil, "optional selector for Helm charts to update")
	flag.Parse()

	if *path == "" {
		fmt.Println("path flag must be set")
		os.Exit(1)
	}
	if !flag.Lookup("provider-selector").Changed {
		providerSelector = nil
	}
	if !flag.Lookup("helm-selector").Changed {
		helmSelector = nil
	}

	// Run update logic
	fs := afero.NewOsFs()
	output, err := update.Update(fs, *path, providerSelector, helmSelector)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(output)
}
