package annotation

import (
	"errors"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	ignoreComment = "#tf-latest-version:ignore"
)

type Annotation struct {
	Content string
	Token   hclsyntax.Token
}

func ParseAnnotations(b []byte) ([]*Annotation, error) {
	aa := []*Annotation{}

	tokens, diags := hclsyntax.LexConfig(b, "main.hcl", hcl.InitialPos)
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
		tokenValue := string(a.Token.Bytes)
		tokenValue = strings.TrimSuffix(tokenValue, "\n")
		if a.Token.Range.Start.Line == r.Start.Line-1 && tokenValue == ignoreComment {
			return true
		}
	}
	return false
}
