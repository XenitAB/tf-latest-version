package result

import (
	"bytes"
	"fmt"
	"text/template"
)

type Update struct {
	Name       string
	OldVersion string
	NewVersion string
}

type Ignore struct {
	Name string
	Path string
}

type Result struct {
	Title   string
	Ignored []Ignore
	Updated []Update
}

func NewResult(title string) Result {
	return Result{
		Title:   title,
		Ignored: []Ignore{},
		Updated: []Update{},
	}
}

func filterUnique(res Result) Result {
	existingUpdated := map[string]string{}
	updated := []Update{}
	for _, u := range res.Updated {
		v, ok := existingUpdated[u.Name]
		// result already in list
		if ok && v == u.NewVersion {
			continue
		}

		existingUpdated[u.Name] = u.NewVersion
		updated = append(updated, u)
	}
	res.Updated = updated

	existingIgnored := map[string]string{}
	ignored := []Ignore{}
	for _, u := range res.Ignored {
		v, ok := existingIgnored[u.Name]
		// result already in list
		if ok && v == u.Path {
			continue
		}

		existingIgnored[u.Name] = u.Path
		ignored = append(ignored, u)
	}
	res.Ignored = ignored

	return res
}

func (r Result) ToMarkdown() (string, error) {
	res := filterUnique(r)
	if len(res.Updated) == 0 && len(res.Ignored) == 0 {
		return fmt.Sprintf("# %s\nNo Changes.", r.Title), nil
	}

	tmpl, err := template.New("markdown").Parse(mdTemplate)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, res)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

const mdTemplate = `# {{ .Title }}
{{- if .Updated }}
## Updated
| Name | Old Version | New Version |
| --- | --- | --- |
{{- range .Updated }}
| {{ .Name }} | {{ .OldVersion }} | {{ .NewVersion }} |
{{- end }}
{{- end }}

{{- if .Ignored }}
## Ignored
| Name | Path |
| --- | --- |
{{- range .Ignored }}
| {{ .Name }} | {{ .Path }} |
{{- end -}}
{{- end -}}
`
