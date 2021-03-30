package helm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func createFs(content string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("/tmp/terraform/", os.FileMode(777))
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

func readFs(fs afero.Fs) (string, error) {
	file, err := fs.Open("/tmp/terraform/main.tf")
	if err != nil {
		return "", err
	}
	d, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

func TestBasic(t *testing.T) {
	fs, err := createFs(basicTerraform)
	assert.Nil(t, err)

	results, err := Update(fs, "/tmp/terraform/")
	assert.Nil(t, err)

	assert.NotEmpty(t, results, "result list can not be empty")
	assert.Equal(t, "aad-pod-identity", results[0].Name)
	assert.Equal(t, "3.0.3", results[0].Version)

	d, err := readFs(fs)
	assert.Nil(t, err)
	assert.Equal(t, basicTerraformExpected, d)
}

func TestInvalidChart(t *testing.T) {
	fs, err := createFs(invalidChartTerraform)
	assert.Nil(t, err)

	_, err = Update(fs, "/tmp/terraform/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find chart entry")
}

const basicTerraform = `
resource "helm_release" "aad_pod_identity" {
  repository = "https://raw.githubusercontent.com/Azure/aad-pod-identity/master/charts"
  chart      = "aad-pod-identity"
  name       = "aad-pod-identity"
  version    = "2.1.0"
}
`

const basicTerraformExpected = `
resource "helm_release" "aad_pod_identity" {
  repository = "https://raw.githubusercontent.com/Azure/aad-pod-identity/master/charts"
  chart      = "aad-pod-identity"
  name       = "aad-pod-identity"
  version    = "3.0.3"
}
`
const invalidChartTerraform = `
resource "helm_release" "aad_pod_identity" {
  repository = "https://raw.githubusercontent.com/Azure/aad-pod-identity/master/charts"
  chart      = "foobar"
  name       = "aad-pod-identity"
  version    = "2.1.0"
}
`
