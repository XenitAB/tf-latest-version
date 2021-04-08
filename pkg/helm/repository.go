package helm

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type Repository interface {
	getLatestVersion(URL, chart string) (string, error)
}

type HelmRepository struct {
	cache map[string]string
}

func NewHelmRepository() HelmRepository {
	return HelmRepository{
		cache: map[string]string{},
	}
}

func (h HelmRepository) getLatestVersion(URL, chart string) (string, error) {
	cacheKey := fmt.Sprintf("%s/%s", URL, chart)
	if v, ok := h.cache[cacheKey]; ok {
		return v, nil
	}

	httpGetter := getter.Provider{
		Schemes: []string{"https", "http"},
		New: func(options ...getter.Option) (getter.Getter, error) {
			return getter.NewHTTPGetter(options...)
		},
	}
	entry := repo.Entry{
		URL: URL,
	}
	chartRepository, err := repo.NewChartRepository(&entry, getter.Providers{httpGetter})
	if err != nil {
		return "", err
	}

	path, err := chartRepository.DownloadIndexFile()
	if err != nil {
		return "", err
	}
	indexFile, err := repo.LoadIndexFile(path)
	if err != nil {
		return "", err
	}
	indexFile.SortEntries()

	chartVersions, ok := indexFile.Entries[chart]
	if !ok {
		return "", fmt.Errorf("could not find chart entry %q", chart)
	}

	if len(chartVersions) == 0 {
		return "", fmt.Errorf("chart %q does not have any versions", chart)
	}

	v := chartVersions[0].Version
	h.cache[cacheKey] = v
	return v, nil
}

type fakeRepository struct {
	charts map[string]repo.ChartVersions
}

func (f fakeRepository) getLatestVersion(URL, chart string) (string, error) {
	chartVersion, ok := f.charts[chart]
	if !ok {
		return "", fmt.Errorf("could not find chart entry %q", chart)
	}

	return firstStableVersion(chartVersion)
}

func firstStableVersion(chartVersions repo.ChartVersions) (string, error) {
	for _, ch := range chartVersions {
		v, err := semver.NewVersion(ch.Version)
		if err != nil {
			return "", fmt.Errorf("could not parse semver %q: %v", ch.Version, err)
		}

		if v.Prerelease() != "" {
			continue
		}

		return ch.Version, nil
	}

	return "", errors.New("no stable versions found")
}
