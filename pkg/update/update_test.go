package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	results := []Result{
		{
			Name:    "foo",
			Version: "1",
		},
		{
			Name:    "bar",
			Version: "2",
		},
	}

	md, err := ToMarkdown("test", results)
	assert.NoError(t, err)
	assert.Equal(t, basicResult, md)
}

const basicResult = `# test
| Name | Version |
| --- | --- |
| foo | 1 |
| bar | 2 |`
