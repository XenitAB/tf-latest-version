package provider

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/pkg/update"
)

func Update(fs afero.Fs, path string) ([]update.Result, error) {
	results := []update.Result{}

	m, diag := tfconfig.LoadModuleFromFilesystem(fsShim{fs}, path)
	if diag.HasErrors() {
		return nil, errors.New(diag.Error())
	}
	for k, p := range m.RequiredProviders {
		latestVersion, err := getLatestVersion(p.Source)
		if err != nil {
			return nil, err
		}
		o, err := tfupdate.NewOption("provider", k, latestVersion, false, []string{})
		if err != nil {
			return nil, err
		}
		err = tfupdate.UpdateFileOrDir(fs, path, o)
		if err != nil {
			return nil, err
		}
		results = append(results, update.Result{Name: p.Source, Version: latestVersion})
	}

	return results, nil
}

type fsShim struct {
	fs afero.Fs
}

func (f fsShim) Open(name string) (tfconfig.File, error) {
	return f.fs.Open(name)
}

func (f fsShim) ReadFile(name string) ([]byte, error) {
	file, err := f.fs.Open(name)
	if err != nil {
		return []byte{}, err
	}
	d, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte{}, err
	}
	return d, err
}

func (f fsShim) ReadDir(dirname string) ([]os.FileInfo, error) {
	file, err := f.fs.Open(dirname)
	if err != nil {
		return []os.FileInfo{}, err
	}
	defer file.Close()
	return file.Readdir(0)
}
