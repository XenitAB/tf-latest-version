package provider

import (
	"errors"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/internal/annotation"
	"github.com/xenitab/tf-provider-latest/internal/result"
)

func Update(fs afero.Fs, path string, r Registry) (*result.Result, error) {
	// Parse the file contents
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	d, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	file.Close()
	hclFile, diags := hclsyntax.ParseConfig(d, path, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}
	pp, err := parseRequiredProviders(hclFile)
	if err != nil {
		return nil, err
	}
	aa, err := annotation.ParseAnnotations(string(d))
	if err != nil {
		return nil, err
	}

	// Loop all of the providers
	res := result.NewResult("Provider")
	for _, p := range pp {
		if annotation.ShouldSkipBlock(aa, p.blockRange) {
			res.Ignored = append(res.Ignored, &result.Ignore{Name: p.source, Path: path})
			continue
		}

		latestVersion, err := r.getLatestVersion(p.source)
		if err != nil {
			return nil, err
		}
		if latestVersion == p.version {
			continue
		}

		o, err := tfupdate.NewOption("provider", p.name, latestVersion, false, []string{})
		if err != nil {
			return nil, err
		}
		err = tfupdate.UpdateFileOrDir(fs, path, o)
		if err != nil {
			return nil, err
		}
		res.Updated = append(res.Updated, &result.Update{Name: p.source, OldVersion: p.version, NewVersion: latestVersion})
	}

	return res, nil
}

type provider struct {
	name       string
	source     string
	version    string
	blockRange hcl.Range
}

func parseRequiredProviders(file *hcl.File) ([]*provider, error) {
	rootSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "terraform",
				LabelNames: nil,
			},
		},
	}
	requiredProvidersSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "required_providers",
				LabelNames: nil,
			},
		},
	}
	rootContent, _, diags := file.Body.PartialContent(rootSchema)
	if diags.HasErrors() {
		return []*provider{}, errors.New(diags.Error())
	}

	pp := []*provider{}
	for _, block := range rootContent.Blocks {
		if block.Type != "terraform" {
			continue
		}

		requiredProvidersContent, _, diags := block.Body.PartialContent(requiredProvidersSchema)
		if diags.HasErrors() {
			return []*provider{}, errors.New(diags.Error())
		}

		// skipping if no required_providers are set
		if len(requiredProvidersContent.Blocks) == 0 {
			continue
		}

		attrs, diags := requiredProvidersContent.Blocks[0].Body.JustAttributes()
		if diags.HasErrors() {
			return []*provider{}, errors.New(diags.Error())
		}

		for k, v := range attrs {
			value, diags := v.Expr.Value(nil)
			if diags.HasErrors() {
				return []*provider{}, errors.New(diags.Error())
			}

			mapValues := value.AsValueMap()
			version := mapValues["version"].AsString()
			source := mapValues["source"].AsString()
			pp = append(pp, &provider{
				name:       k,
				version:    version,
				source:     source,
				blockRange: v.Range,
			})
		}
	}

	return pp, nil
}
