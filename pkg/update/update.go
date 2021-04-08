package update

import (
	"bytes"
	"text/template"
)

type Result struct {
	Name    string
	Version string
}

func UniqueResults(rr []Result) []Result {
	existing := map[string]string{}
	results := []Result{}
	for _, r := range rr {
		v, ok := existing[r.Name]
		// result already in list
		if ok && v == r.Version {
			continue
		}

		existing[r.Name] = r.Version
		results = append(results, r)
	}
	return results
}

func ToMarkdown(title string, results []Result) (string, error) {
	tmpl, err := template.New("markdown").Parse(mdTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		Title   string
		Results []Result
	}{
		Title:   title,
		Results: results,
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

const mdTemplate = `# {{ .Title }}
| Name | Version |
| --- | --- |
{{- range .Results }}
| {{ .Name }} | {{ .Version }} |
{{- end -}}
`
