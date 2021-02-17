package scripts

import (
	"embed"
	"io/fs"
)

//go:embed *.sh
var content embed.FS

func Open(path string) (fs.File, error) {
	return content.Open(path)
}
