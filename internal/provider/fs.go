package provider

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/afero"
)

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
