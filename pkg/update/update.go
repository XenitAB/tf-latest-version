package update

import (
	"bytes"
	"text/template"
)

type Result struct {
	Name    string
	Version string
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
