package provider

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func createFs(content string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("/tmp/terraform/", os.FileMode(0777))
	if err != nil {
		return nil, err
	}
	f, err := fs.Create("/tmp/terraform/main.tf")
	if err != nil {
		return nil, err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func TestProviderBasic(t *testing.T) {
	fs, err := createFs(basicTerraform)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.53.0"},
		},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r)
	require.Nil(t, err)
	require.NotEmpty(t, res.Updated, "result list can not be empty")
	require.Equal(t, "hashicorp/azurerm", res.Updated[0].Name)
	require.Equal(t, "2.53.0", res.Updated[0].NewVersion)

	file, err := fs.Open("/tmp/terraform/main.tf")
	require.Nil(t, err)
	d, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, basicTerraformExpected, string(d))
}

func TestProviderEmptyRequired(t *testing.T) {
	fs, err := createFs(noRequiredProviders)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{},
	}
	_, err = Update(fs, "/tmp/terraform/main.tf", r)
	require.Nil(t, err)
}

func TestProviderIgnore(t *testing.T) {
	fs, err := createFs(ignoreTerraform)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.53.0"},
		},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r)
	require.Nil(t, err)
	require.Empty(t, res.Updated)
	require.NotEmpty(t, res.Ignored)
}

func TestProviderFalsePositive(t *testing.T) {
	fs, err := createFs(falsePositiveTerraform)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.53.0"},
		},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r)
	require.Nil(t, err)
	require.NotEmpty(t, res.Updated)
	require.Empty(t, res.Ignored)
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

const ignoreTerraform = `
terraform {
  required_version = "0.13.5"

  required_providers {
		#tf-latest-version:ignore
    azurerm = {
      source  = "hashicorp/azurerm"
			version = "2.36.0"
    }
  }
}

provider "azurerm" {}
`

const falsePositiveTerraform = `
terraform {
  required_version = "0.13.5"

  required_providers {
		#do-not:ignore
    azurerm = {
      source  = "hashicorp/azurerm"
			version = "2.36.0"
    }
  }
}

provider "azurerm" {}
`
