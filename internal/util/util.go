package util

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/internal/annotation"
)

//nolint:gocritic // skip as it just makes things more verbose
func ReadHCLFile(fs afero.Fs, path string) (*hcl.File, *hclwrite.File, []*annotation.Annotation, error) {
	b, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, nil, nil, err
	}
	hclWriteFile, diags := hclwrite.ParseConfig(b, "main.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, nil, nil, errors.New(diags.Error())
	}
	hclFile, diags := hclsyntax.ParseConfig(b, path, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, nil, nil, errors.New(diags.Error())
	}
	annos, err := annotation.ParseAnnotations(b)
	if err != nil {
		return nil, nil, nil, err
	}
	return hclFile, hclWriteFile, annos, nil
}

func ReplaceHCLFile(fs afero.Fs, path string, hclFile *hclwrite.File) error {
	err := fs.Remove(path)
	if err != nil {
		return err
	}
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = hclFile.WriteTo(file)
	if err != nil {
		return err
	}
	return nil
}
