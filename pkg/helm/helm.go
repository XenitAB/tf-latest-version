package helm

import (
	"errors"
	"fmt"
	iofs "io/fs"
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

func Update(fs afero.Fs, r Repository, path string) ([]update.Result, error) {
	results := []update.Result{}
	err := afero.Walk(fs, path, func(path string, info iofs.FileInfo, err error) error {
		// Ignore directories
		if info.IsDir() {
			return nil
		}

		// Skip non Terraform files
		if filepath.Ext(info.Name()) != terraformExtension {
			return nil
		}

		// Read and parse the content of the file
		file, err := fs.Open(path)
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
			attrs := block.Body().Attributes()
			chart, err := attributeString(attrs, "chart")
			if err != nil {
				return err
			}
			repository, err := attributeString(attrs, "repository")
			if err != nil {
				// skip if the repository is not set
				continue
			}
			version, err := attributeString(attrs, "version")
			if err != nil {
				// ok if version is not set
				version = ""
			}
			latestVersion, err := r.getLatestVersion(repository, chart)
			if err != nil {
				return err
			}
			// skip if version is already latest
			if version == latestVersion {
				continue
			}

			// Update the block with the latest version
			block.Body().SetAttributeValue("version", cty.StringVal(latestVersion))
			results = append(results, update.Result{Name: chart, Version: latestVersion})
		}

		// Clear the old file and write the new content
		err = fs.Remove(path)
		if err != nil {
			return err
		}
		file, err = fs.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = hclFile.WriteTo(file)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
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
