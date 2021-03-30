package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func getLatestVersion(URL, chart string) (string, error) {
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

	return chartVersions[0].Version, nil
}
