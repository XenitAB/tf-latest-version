package update

import (
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/internal/helm"
	"github.com/xenitab/tf-provider-latest/internal/provider"
	"github.com/xenitab/tf-provider-latest/internal/result"
)

const TerraformExtension = ".tf"

func Update(fs afero.Fs, path string) (string, error) {
	resMap := map[string]result.Result{}

	err := afero.Walk(fs, path, func(path string, info iofs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(info.Name()) != TerraformExtension {
			return nil
		}

		helmResult, err := helm.Update(fs, path, helm.NewHelmRepository())
		if err != nil {
			return err
		}
		resMap = merge(resMap, helmResult)
		providerResult, err := provider.Update(fs, path, provider.NewHashicorpRegistry())
		if err != nil {
			return err
		}
		resMap = merge(resMap, providerResult)

		return nil
	})
	if err != nil {
		return "", err
	}

	outputs := []string{}
	for _, r := range resMap {
		output, err := r.ToMarkdown()
		if err != nil {
			return "", err
		}
		outputs = append(outputs, output)
	}
	return strings.Join(outputs, "\n\n"), nil
}

func merge(resMap map[string]result.Result, res result.Result) map[string]result.Result {
	exist, ok := resMap[res.Title]
	if !ok {
		resMap[res.Title] = res
		return resMap
	}

	exist.Updated = append(exist.Updated, res.Updated...)
	exist.Ignored = append(exist.Ignored, res.Ignored...)
	resMap[res.Title] = exist
	return resMap
}
