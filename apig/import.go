package apig

import (
	"path/filepath"
	"strings"
)

func formatImportDir(paths []string) []string {
	results := make([]string, 0, len(paths))
	flag := map[string]bool{}
	for i := 0; i < len(paths); i++ {
		dir := filepath.Dir(paths[i])
		if !flag[dir] && dir != "." {
			flag[dir] = true
			results = append(results, dir)
		}
	}

	if len(results) > 1 {
		//Naive approache is to find one with /db, in the future if this doesn't work
		//we can find one that matches the current path
		for _, ipath := range paths {
			if strings.Index(ipath, "/db") > 0 {
				return []string{filepath.Dir(ipath)}
			}
		}
	}

	return results
}
