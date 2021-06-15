package config

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
)

func transform(vs []string) []cty.Value {
	vsm := make([]cty.Value, len(vs))
	for i, v := range vs {
		vsm[i] = cty.StringVal(v)
	}
	return vsm
}

func makeAbsolute(path string, base string) string {
	_, filename := filepath.Split(expandPath(path))
	return base + "/" + filename
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}
