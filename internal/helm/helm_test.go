package helm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
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
	require.Nil(t, err)

	r := fakeRepository{
		charts: map[string]repo.ChartVersions{
			"aad-pod-identity": {
				{
					Metadata: &chart.Metadata{
						Version: "3.0.3",
					},
				},
			},
		},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r)
	require.Nil(t, err)

	require.NotEmpty(t, res.Updated, "result list can not be empty")
	require.Equal(t, "aad-pod-identity", res.Updated[0].Name)
	require.Equal(t, "3.0.3", res.Updated[0].NewVersion)

	d, err := readFs(fs)
	require.Nil(t, err)
	require.Equal(t, basicTerraformExpected, d)
}

func TestInvalidChart(t *testing.T) {
	fs, err := createFs(invalidChartTerraform)
	require.Nil(t, err)

	r := fakeRepository{
		charts: map[string]repo.ChartVersions{
			"aad-pod-identity": {
				{
					Metadata: &chart.Metadata{
						Version: "3.0.3",
					},
				},
			},
		},
	}
	_, err = Update(fs, "/tmp/terraform/main.tf", r)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not find chart entry")
}

func TestIgnoreChart(t *testing.T) {
	fs, err := createFs(ignoreTerraform)
	require.Nil(t, err)

	r := fakeRepository{
		charts: map[string]repo.ChartVersions{
			"aad-pod-identity": {
				{
					Metadata: &chart.Metadata{
						Version: "3.0.3",
					},
				},
			},
		},
	}
	res, err := Update(fs, "/tmp/terraform/main.tf", r)
	require.Nil(t, err)
	require.Empty(t, res.Updated)
	require.NotEmpty(t, res.Ignored)
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

const ignoreTerraform = `
#tf-latest-version:ignore
resource "helm_release" "aad_pod_identity" {
  repository = "https://raw.githubusercontent.com/Azure/aad-pod-identity/master/charts"
  chart      = "aad-pod-identity"
  name       = "aad-pod-identity"
	version    = "2.1.0"
}
`
