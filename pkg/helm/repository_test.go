package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

func TestFirstStableVersion(t *testing.T) {
	chartVersions := repo.ChartVersions{
		{
			Metadata: &chart.Metadata{
				Version: "0.0.1-rc1",
			},
		},
		{
			Metadata: &chart.Metadata{
				Version: "0.0.1",
			},
		},
		{
			Metadata: &chart.Metadata{
				Version: "0.0.1-beta1",
			},
		},
	}

	v, err := firstStableVersion(chartVersions)
	require.NoError(t, err)
	require.Equal(t, "0.0.1", v)
}

func TestFirstStableVersionNone(t *testing.T) {
	chartVersions := repo.ChartVersions{
		{
			Metadata: &chart.Metadata{
				Version: "0.0.1-rc1",
			},
		},
		{
			Metadata: &chart.Metadata{
				Version: "0.0.1-foo",
			},
		},
		{
			Metadata: &chart.Metadata{
				Version: "0.0.1-beta1",
			},
		},
	}

	_, err := firstStableVersion(chartVersions)
	require.Error(t, err)
}
