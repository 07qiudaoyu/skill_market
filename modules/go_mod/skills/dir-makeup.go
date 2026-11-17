package skills

import (
	"path/filepath"
)

func Dirmakeup(dir string) string {
	var ck bool
	var dir1 string
	if dir != "" {
		ck = true

	} else {
		ck = false
	}
	if ck == false {
		dir1 = filepath.Join("..", "..", "download-zip")
	} else {
		dir1 = filepath.Join(dir)
	}
	return dir1
}
