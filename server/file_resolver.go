package server

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"java-mini-ls-go/util"
	"path/filepath"
	"strings"
)

// FileResolver represents any file system operations that we'd want to be able to mock out for unit tests.
type FileResolver interface {
	FileURIToPath(uri string) (string, error)
	ListJavaFilesRecursive(folderPath string) ([]string, error)
	ReadFile(filePath string) string
}

type RealFileResolver struct {
	log *zap.Logger
}

var _ FileResolver = (*RealFileResolver)(nil)

func (r *RealFileResolver) FileURIToPath(uri string) (string, error) {
	return util.FileURIToPath(uri)
}

func (r *RealFileResolver) ListJavaFilesRecursive(filePath string) ([]string, error) {
	results, err := ioutil.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, f := range results {
		if f.IsDir() {
			// Skip hidden folders like .git
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}

			subDir := filepath.Join(filePath, f.Name())
			subDirContents, err := r.ListJavaFilesRecursive(subDir)
			if err != nil {
				return nil, err
			}
			ret = append(ret, subDirContents...)
		} else if strings.HasSuffix(f.Name(), ".java") {
			ret = append(ret, filepath.Join(filePath, f.Name()))
		}
	}

	return ret, nil
}

func (r *RealFileResolver) ReadFile(filePath string) string {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		err = errors.Wrapf(err, "error reading file at %s", filePath)
		r.log.Error(err.Error())
		return ""
	}

	return string(ret)
}
