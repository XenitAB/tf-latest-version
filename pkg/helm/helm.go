package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"

	"github.com/xenitab/tf-provider-latest/pkg/update"
)

const terraformExtension = ".tf"

func Update(fs afero.Fs, path string) ([]update.Result, error) {
	results := []update.Result{}

	// List all of the files in the path
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		d, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		file.Close()
		hclFile, diag := hclwrite.ParseConfig(d, "main.hcl", hcl.InitialPos)
		if diag.HasErrors() {
			return nil, errors.New(diag.Error())
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
			attrs := block.Body().Attributes()
			chart, err := attributeString(attrs, "chart")
			if err != nil {
				return nil, err
			}
			repository, err := attributeString(attrs, "repository")
			if err != nil {
				// skip if the repository is not set
				continue
			}
			latestVersion, err := getLatestVersion(repository, chart)
			if err != nil {
				return nil, err
			}

			// Update the block with the latest version
			block.Body().SetAttributeValue("version", cty.StringVal(latestVersion))
			results = append(results, update.Result{Name: chart, Version: latestVersion})
		}

		// Clear the old file and write the new content
		err = fs.Remove(filePath)
		if err != nil {
			return nil, err
		}
		file, err = fs.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		_, err = hclFile.WriteTo(file)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func attributeString(attrs map[string]*hclwrite.Attribute, key string) (string, error) {
	attr, ok := attrs[key]
	if !ok {
		return "", fmt.Errorf("could not get attribute for key %q", key)
	}

	s := string(attr.Expr().BuildTokens(nil).Bytes())
	s = strings.Trim(s, " \"")
	s = strings.Trim(s, "\"")
	return s, nil
}
