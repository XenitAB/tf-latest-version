package helm

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"

	"github.com/xenitab/tf-provider-latest/internal/annotation"
	"github.com/xenitab/tf-provider-latest/internal/result"
	"github.com/xenitab/tf-provider-latest/internal/util"
)

func Update(fs afero.Fs, path string, r Repository, helmSelector *[]string) (*result.Result, error) {
	hclFile, hclWriteFile, annos, err := util.ReadHCLFile(fs, path)
	if err != nil {
		return nil, err
	}
	hh, err := parseHelmReleases(hclFile)
	if err != nil {
		return nil, err
	}

	selector := map[string]string{}
	if helmSelector != nil {
		for _, s := range *helmSelector {
			selector[s] = s
		}
	}
	res := result.NewResult("Helm")
	for _, h := range hh {
		// Skip if the repository is not set as it mean the chart is local
		if h.repository == "" {
			continue
		}

		if _, ok := selector[h.chart]; helmSelector != nil && !ok {
			res.Ignored = append(res.Ignored, &result.Ignore{Name: h.chart, Path: path})
			continue
		}
		if annotation.ShouldSkipBlock(annos, h.blockRange) {
			res.Ignored = append(res.Ignored, &result.Ignore{Name: h.chart, Path: path})
			continue
		}

		latestVersion, err := r.getLatestVersion(h.repository, h.chart)
		if err != nil {
			res.Failed = append(res.Failed, &result.Failure{
				Name:    h.chart,
				Path:    path,
				Message: err.Error(),
			})
			continue
		}
		if h.version == latestVersion {
			continue
		}

		block := hclWriteFile.Body().FirstMatchingBlock("resource", []string{"helm_release", h.name})
		if block == nil {
			res.Failed = append(res.Failed, &result.Failure{
				Name:    h.chart,
				Path:    path,
				Message: "block cannot be nil",
			})
			continue
		}
		block.Body().SetAttributeValue("version", cty.StringVal(latestVersion))
		res.Updated = append(res.Updated, &result.Update{
			Name:       h.chart,
			OldVersion: h.version,
			NewVersion: latestVersion,
		})
	}

	err = util.ReplaceHCLFile(fs, path, hclWriteFile)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type helmRelease struct {
	name       string
	version    string
	chart      string
	repository string
	blockRange hcl.Range
}

type helmReleaseResource struct {
	Version    string   `hcl:"version,optional"`
	Chart      string   `hcl:"chart"`
	Repository string   `hcl:"repository,optional"`
	Remain     hcl.Body `hcl:",remain"`
}

func parseHelmReleases(file *hcl.File) ([]*helmRelease, error) {
	rootSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
		},
	}
	content, _, diags := file.Body.PartialContent(rootSchema)
	if diags.HasErrors() {
		return []*helmRelease{}, errors.New(diags.Error())
	}

	hh := []*helmRelease{}
	for _, block := range content.Blocks {
		if block.Type != "resource" {
			continue
		}

		if len(block.Labels) == 0 {
			continue
		}

		if block.Labels[0] != "helm_release" {
			continue
		}

		var hrr helmReleaseResource
		ctx := &hcl.EvalContext{
			Variables: map[string]cty.Value{
				"path": cty.MapVal(map[string]cty.Value{
					"module": cty.StringVal("foobar"),
				}),
			},
		}
		diags := gohcl.DecodeBody(block.Body, ctx, &hrr)
		if diags.HasErrors() {
			return []*helmRelease{}, errors.New(diags.Error())
		}

		hh = append(hh, &helmRelease{
			name:       block.Labels[1],
			version:    hrr.Version,
			chart:      hrr.Chart,
			repository: hrr.Repository,
			blockRange: block.DefRange,
		})
	}

	return hh, nil
}
