package provider

import (
	"errors"
	iofs "io/fs"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/minamijoyo/tfupdate/tfupdate"
	"github.com/spf13/afero"

	"github.com/xenitab/tf-provider-latest/pkg/update"
)

func Update(fs afero.Fs, r Registry, path string) ([]update.Result, error) {
	results := []update.Result{}

	err := afero.Walk(fs, path, func(path string, info iofs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		m, diag := tfconfig.LoadModuleFromFilesystem(fsShim{fs}, path)
		if diag.HasErrors() {
			return errors.New(diag.Error())
		}

		for k, p := range m.RequiredProviders {
			if p.Source == "" {
				continue
			}
			latestVersion, err := r.getLatestVersion(p.Source)
			if err != nil {
				return err
			}
			o, err := tfupdate.NewOption("provider", k, latestVersion, false, []string{})
			if err != nil {
				return err
			}
			err = tfupdate.UpdateFileOrDir(fs, path, o)
			if err != nil {
				return err
			}
			results = append(results, update.Result{Name: p.Source, Version: latestVersion})
		}

		return nil
	})
	if err != nil {
		return nil, err
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
