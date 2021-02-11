package helm

import (
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
	return indexFile.Entries[chart][0].Version, nil
}
