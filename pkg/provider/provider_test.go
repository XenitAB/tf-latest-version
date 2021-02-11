package provider

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("/tmp/terraform/", os.FileMode(777))
	assert.Nil(t, err)
	f, err := fs.Create("/tmp/terraform/main.tf")
	assert.Nil(t, err)
	n, err := f.WriteString(basicTerraform)
	assert.Nil(t, err)
	assert.Equal(t, len(basicTerraform), n)
	err = f.Close()
	assert.Nil(t, err)

	err = Update(fs, "/tmp/terraform/")
	assert.Nil(t, err)

	file, err := fs.Open("/tmp/terraform/main.tf")
	assert.Nil(t, err)
	d, err := ioutil.ReadAll(file)
	assert.Nil(t, err)
	assert.Equal(t, basicTerraformExpected, string(d))
}

const basicTerraform = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.35.0"
    }
  }
}
`

const basicTerraformExpected = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.46.1"
    }
  }
}
`
