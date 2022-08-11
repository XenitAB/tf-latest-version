package result

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	res := Result{
		Title: "test",
		Updated: []*Update{
			{
				Name:       "foo",
				OldVersion: "0",
				NewVersion: "1",
			},
		},
		Ignored: []*Ignore{
			{
				Name: "bar",
				Path: "baz",
			},
		},
		Failed: []*Failure{
			{
				Name:    "foo",
				Path:    "bar",
				Message: "baz",
			},
		},
	}

	md, err := res.ToMarkdown()
	assert.NoError(t, err)
	assert.Equal(t, allResult, md)
}

func TestUpdated(t *testing.T) {
	res := Result{
		Title: "test",
		Updated: []*Update{
			{
				Name:       "foo",
				OldVersion: "0",
				NewVersion: "1",
			},
			{
				Name:       "bar",
				OldVersion: "1",
				NewVersion: "2",
			},
			{
				Name:       "bar",
				OldVersion: "1",
				NewVersion: "2",
			},
		},
		Ignored: []*Ignore{},
		Failed:  []*Failure{},
	}

	md, err := res.ToMarkdown()
	assert.NoError(t, err)
	assert.Equal(t, updatedResult, md)
}

func TestIgnored(t *testing.T) {
	res := Result{
		Title:   "test",
		Updated: []*Update{},
		Ignored: []*Ignore{
			{
				Name: "bar",
				Path: "baz",
			},
			{
				Name: "bar",
				Path: "baz",
			},
		},
		Failed: []*Failure{},
	}

	md, err := res.ToMarkdown()
	assert.NoError(t, err)
	assert.Equal(t, ignoredResult, md)
}

func TestFailed(t *testing.T) {
	res := Result{
		Title:   "test",
		Updated: []*Update{},
		Ignored: []*Ignore{},
		Failed: []*Failure{
			{
				Name:    "foo",
				Path:    "bar",
				Message: "baz",
			},
			{
				Name:    "foo",
				Path:    "bar",
				Message: "baz",
			},
		},
	}

	md, err := res.ToMarkdown()
	assert.NoError(t, err)
	assert.Equal(t, failedResult, md)
}

func TestNone(t *testing.T) {
	res := Result{
		Title:   "test",
		Updated: []*Update{},
		Ignored: []*Ignore{},
		Failed:  []*Failure{},
	}

	md, err := res.ToMarkdown()
	assert.NoError(t, err)
	assert.Equal(t, noneResult, md)
}

const allResult = `# test
## Updated
| Name | Old Version | New Version |
| --- | --- | --- |
| foo | 0 | 1 |
## Ignored
| Name | Path |
| --- | --- |
| bar | baz |
## Failed
| Name | Path | Message |
| --- | --- | --- |
| foo | bar | baz |`

const updatedResult = `# test
## Updated
| Name | Old Version | New Version |
| --- | --- | --- |
| foo | 0 | 1 |
| bar | 1 | 2 |`

const ignoredResult = `# test
## Ignored
| Name | Path |
| --- | --- |
| bar | baz |`

const failedResult = `# test
## Failed
| Name | Path | Message |
| --- | --- | --- |
| foo | bar | baz |`

const noneResult = `# test
No Changes.`
