package helm

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

func Update(fs afero.Fs, path string) error {
	// Read files in the path
	file, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	files, err := file.Readdir(0)
	if err != nil {
		return err
	}

	// Iterate through all files
	for _, fileInfo := range files {
		if filepath.Ext(fileInfo.Name()) != "tf" {
			continue
		}

		file, err := fs.Open(fileInfo.Name())
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

		for _, block := range hclFile.Body().Blocks() {
			if block.Labels()[0] != "helm_release" {
				continue
			}
			attr := block.Body().Attributes()
			chart := cleanString(string(attr["chart"].Expr().BuildTokens(nil).Bytes()))
			repository := cleanString(string(attr["repository"].Expr().BuildTokens(nil).Bytes()))
			latestVersion, err := getLatestVersion(repository, chart)
			if err != nil {
				return err
			}
			block.Body().SetAttributeValue("version", cty.StringVal(latestVersion))
		}

		file, err = fs.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
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
