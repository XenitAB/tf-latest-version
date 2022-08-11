package provider

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"

	"github.com/xenitab/tf-provider-latest/internal/annotation"
	"github.com/xenitab/tf-provider-latest/internal/result"
	"github.com/xenitab/tf-provider-latest/internal/util"
)

func Update(fs afero.Fs, path string, reg Registry, providerSelector *[]string) (*result.Result, error) {
	hclFile, _, annos, err := util.ReadHCLFile(fs, path)
	if err != nil {
		return nil, err
	}
	pp, err := parseRequiredProviders(hclFile)
	if err != nil {
		return nil, err
	}

	selector := map[string]string{}
	if providerSelector != nil {
		for _, s := range *providerSelector {
			selector[s] = s
		}
	}
	res := result.NewResult("Provider")
	for _, p := range pp {
		if _, ok := selector[p.source]; providerSelector != nil && !ok {
			res.Ignored = append(res.Ignored, &result.Ignore{Name: p.source, Path: path})
			continue
		}
		if annotation.ShouldSkipBlock(annos, p.blockRange) {
			res.Ignored = append(res.Ignored, &result.Ignore{Name: p.source, Path: path})
			continue
		}

		latestVersion, err := reg.getLatestVersion(p.source)
		if err != nil {
			res.Failed = append(res.Failed, &result.Failure{
				Name:    p.source,
				Path:    path,
				Message: err.Error(),
			})
			continue
		}
		if latestVersion == p.version {
			continue
		}

		o, err := tfupdate.NewOption("provider", p.name, latestVersion, false, []string{})
		if err != nil {
			res.Failed = append(res.Failed, &result.Failure{
				Name:    p.source,
				Path:    path,
				Message: err.Error(),
			})
			continue
		}
		err = tfupdate.UpdateFileOrDir(fs, path, o)
		if err != nil {
			res.Failed = append(res.Failed, &result.Failure{
				Name:    p.source,
				Path:    path,
				Message: err.Error(),
			})
			continue
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

		for name, attr := range attrs {
			p, err := parseProvider(name, attr)
			if err != nil {
				return []*provider{}, err
			}
			pp = append(pp, p)
		}
	}

	return pp, nil
}

func parseProvider(name string, attr *hcl.Attribute) (*provider, error) {
	keyValuePairs, diags := hcl.ExprMap(attr.Expr)
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	p := &provider{
		name:       name,
		blockRange: attr.Range,
	}
	//nolint:gocritic // ignore for now
	for _, kvp := range keyValuePairs {
		key, diags := kvp.Key.Value(nil)
		if diags.HasErrors() {
			return nil, errors.New(diags.Error())
		}

		if key.Type() != cty.String {
			return nil, errors.New("invalid key type")
		}

		switch key.AsString() {
		case "version":
			version, diags := kvp.Value.Value(nil)
			if diags.HasErrors() {
				return nil, errors.New(diags.Error())
			}
			p.version = version.AsString()
		case "source":
			source, diags := kvp.Value.Value(nil)
			if diags.HasErrors() {
				return nil, errors.New(diags.Error())
			}
			p.source = source.AsString()
		}
	}
	return p, nil
}
