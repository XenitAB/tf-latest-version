package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/pkg/helm"
	"github.com/xenitab/tf-provider-latest/pkg/provider"
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
	err := provider.Update(fs, string(path))
	if err != nil {
		fmt.Printf("failed updating provider versions: %q\n", err)
		os.Exit(1)
	}
	err = helm.Update(fs, string(path))
	if err != nil {
		fmt.Printf("failed updating helm versions: %q\n", err)
		os.Exit(1)
	}
}
