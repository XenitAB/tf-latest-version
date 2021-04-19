package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/xenitab/tf-provider-latest/internal/update"
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
	output, err := update.Update(fs, string(path))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(output)
}
