package annotation

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	IgnoreComment = "#tf-latest-version:ignore"
)

type Annotation struct {
	Content string
	Token   hclsyntax.Token
}

func ParseAnnotations(hclString string) ([]*Annotation, error) {
	aa := []*Annotation{}

	tokens, diags := hclsyntax.LexConfig([]byte(hclString), "main.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return []*Annotation{}, errors.New(diags.Error())
	}

	//nolint:gocritic // ignore for now
	for _, token := range tokens {
		if token.Type != hclsyntax.TokenComment {
			continue
		}

		aa = append(aa, &Annotation{
			Content: string(token.Bytes),
			Token:   token,
		})
	}

	return aa, nil
}

func ShouldSkipBlock(aa []*Annotation, r hcl.Range) bool {
	for _, a := range aa {
		if a.Token.Range.Start.Line == r.Start.Line-1 {
			return true
		}
	}
	return false
}
