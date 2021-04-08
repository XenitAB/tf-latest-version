package provider

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestProviderBasic(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("/tmp/terraform/", os.FileMode(777))
	require.Nil(t, err)
	f, err := fs.Create("/tmp/terraform/main.tf")
	require.Nil(t, err)
	n, err := f.WriteString(basicTerraform)
	require.Nil(t, err)
	require.Equal(t, len(basicTerraform), n)
	err = f.Close()
	require.Nil(t, err)

	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.53.0"},
		},
	}
	results, err := Update(fs, r, "/tmp/terraform/")
	require.Nil(t, err)
	require.NotEmpty(t, results, "result list can not be empty")
	require.Equal(t, "hashicorp/azurerm", results[0].Name)
	require.Equal(t, "2.53.0", results[0].Version)

	file, err := fs.Open("/tmp/terraform/main.tf")
	require.Nil(t, err)
	d, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, basicTerraformExpected, string(d))
}

func TestProviderEmptyRequired(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("/tmp/terraform/", os.FileMode(777))
	require.Nil(t, err)
	f, err := fs.Create("/tmp/terraform/main.tf")
	require.Nil(t, err)
	n, err := f.WriteString(noRequiredProviders)
	require.Nil(t, err)
	require.Equal(t, len(noRequiredProviders), n)
	err = f.Close()
	require.Nil(t, err)

	r := FakeRegistry{
		providers: map[string][]string{},
	}
	_, err = Update(fs, r, "/tmp/terraform/")
	require.Nil(t, err)
	/*require.NotEmpty(t, results, "result list can not be empty")
	require.Equal(t, "hashicorp/azurerm", results[0].Name)
	require.Equal(t, "2.53.0", results[0].Version)

	file, err := fs.Open("/tmp/terraform/main.tf")
	require.Nil(t, err)
	d, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, basicTerraformExpected, string(d))*/
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

provider "azurerm" {}
`

const basicTerraformExpected = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.53.0"
    }
  }
}

provider "azurerm" {}
`

const noRequiredProviders = `
terraform {
}

provider "aws" {
  region = "eu-west-1"
}
`
