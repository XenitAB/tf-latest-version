package helm

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

const terraformExtension = ".tf"

func Update(fs afero.Fs, path string) error {
	// List all of the files in the path
	file, err := fs.Open(path)
	if err != nil {
		return err
	}
	files, err := file.Readdir(0)
	if err != nil {
		return err
	}
	file.Close()

	// Iterate through all files
	for _, fileInfo := range files {
		if filepath.Ext(fileInfo.Name()) != terraformExtension {
			continue
		}

		// Read and parse the content of the file
		filePath := filepath.Join(path, fileInfo.Name())
		file, err := fs.Open(filePath)
		if err != nil {
			return err
		}
		d, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		file.Close()
		hclFile, diag := hclwrite.ParseConfig(d, "main.hcl", hcl.InitialPos)
		if diag.HasErrors() {
			return errors.New(diag.Error())
		}

		// Iterate through all of the HCL blocks
		for _, block := range hclFile.Body().Blocks() {
			// Check if block is a helm_release
			if len(block.Labels()) == 0 {
				continue
			}

			if block.Labels()[0] != "helm_release" {
				continue
			}

			// Parse the fields and get the latest version
			attr := block.Body().Attributes()
			chart := cleanString(string(attr["chart"].Expr().BuildTokens(nil).Bytes()))
			repository := cleanString(string(attr["repository"].Expr().BuildTokens(nil).Bytes()))
			latestVersion, err := getLatestVersion(repository, chart)
			if err != nil {
				return err
			}

			// Update the block with the latest version
			block.Body().SetAttributeValue("version", cty.StringVal(latestVersion))
		}

		// Clear the old file and write the new content
		err = fs.Remove(filePath)
		if err != nil {
			return err
		}
		file, err = fs.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = hclFile.WriteTo(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanString(s string) string {
	s = strings.Trim(s, " \"")
	s = strings.Trim(s, "\"")
	return s
}
