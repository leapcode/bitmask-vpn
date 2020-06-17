package backend

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

func cleanupTempDirs() {
	dirs, _ := filepath.Glob(path.Join(os.TempDir(), "leap-*"))
	for _, d := range dirs {
		log.Println("removing temp dir:", d)
		os.RemoveAll(d)
	}
}
