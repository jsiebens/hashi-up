package config

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
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

func generate(f *hclwrite.File) string {
	str := `# generated with hashi-up

`
	return str + string(f.Bytes())
}
