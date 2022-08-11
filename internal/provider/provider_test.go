package provider

import (
	"io"
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

func TestProviderUpdate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic",
			input:    basicTerraform,
			expected: basicTerraformExpected,
		},
		{
			name:     "extra config",
			input:    extraConfigTerraform,
			expected: extraConfigTerraformExpected,
		},
	}
	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.53.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := createFs(tt.input)
			require.Nil(t, err)
			res, err := Update(fs, "/tmp/terraform/main.tf", r, nil)
			require.Nil(t, err)
			require.NotEmpty(t, res.Updated, "result list can not be empty")
			require.Equal(t, "hashicorp/azurerm", res.Updated[0].Name)
			require.Equal(t, "2.53.0", res.Updated[0].NewVersion)

			file, err := fs.Open("/tmp/terraform/main.tf")
			require.Nil(t, err)
			d, err := io.ReadAll(file)
			require.Nil(t, err)
			require.Equal(t, tt.expected, string(d))
		})
	}
}

func TestProviderEmptyRequired(t *testing.T) {
	fs, err := createFs(noRequiredProviders)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{},
	}
	_, err = Update(fs, "/tmp/terraform/main.tf", r, nil)
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
	res, err := Update(fs, "/tmp/terraform/main.tf", r, nil)
	require.Nil(t, err)
	require.Empty(t, res.Updated)
	require.Empty(t, res.Failed)
	require.NotEmpty(t, res.Ignored)
}

func TestProviderFail(t *testing.T) {
	fs, err := createFs(failTerraform)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r, nil)
	require.Nil(t, err)
	require.Empty(t, res.Updated)
	require.Empty(t, res.Ignored)
	require.NotEmpty(t, res.Failed)
}

func TestProviderFalsePositive(t *testing.T) {
	fs, err := createFs(falsePositiveTerraform)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.53.0"},
		},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r, nil)
	require.Nil(t, err)
	require.NotEmpty(t, res.Updated)
	require.Empty(t, res.Ignored)
	require.Empty(t, res.Failed)
}

func TestProviderSelector(t *testing.T) {
	fs, err := createFs(providerSelector)
	require.Nil(t, err)
	r := FakeRegistry{
		providers: map[string][]string{
			"hashicorp/azurerm": {"2.77.0"},
			"hashicorp/aws":     {"3.59.0"},
		},
	}
	providerSelector := []string{"hashicorp/azurerm"}
	res, err := Update(fs, "/tmp/terraform/main.tf", r, &providerSelector)

	require.Nil(t, err)
	require.Len(t, res.Updated, 1)
	require.Len(t, res.Ignored, 1)

	file, err := fs.Open("/tmp/terraform/main.tf")
	require.Nil(t, err)
	d, err := io.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, providerSelectorExpected, string(d))
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

const extraConfigTerraform = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.35.0"
      configuration_aliases = [azurerm.foobar]
    }
  }
}

provider "azurerm" {}
`

const extraConfigTerraformExpected = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source                = "hashicorp/azurerm"
      version               = "2.53.0"
      configuration_aliases = [azurerm.foobar]
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

const failTerraform = `
terraform {
  required_version = "0.13.5"

  required_providers {
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

const providerSelector = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.76.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "3.58.0"
    }
  }
}

provider "azurerm" {}
`

const providerSelectorExpected = `
terraform {
  required_version = "0.13.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.77.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "3.58.0"
    }
  }
}

provider "azurerm" {}
`
