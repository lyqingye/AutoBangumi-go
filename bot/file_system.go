package bot

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/studio-b12/gowebdav"
)

type LocalFileSystem struct {
}

func (fs LocalFileSystem) ReadDir(dirPath string) ([]os.FileInfo, error) {
	var ret []os.FileInfo
	fi, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, errors.Errorf("%s is no directory", dirPath)
	}
	dirEntry, err := os.ReadDir(dirPath)
	for _, entry := range dirEntry {
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}
		ret = append(ret, fileInfo)
	}
	return ret, nil
}

func (fs LocalFileSystem) WalkDir(dirPath string, callback func(parent string, info os.FileInfo) (bool, error)) error {
	return walkDirInternal(func(s string) ([]os.FileInfo, error) {
		return fs.ReadDir(s)
	}, dirPath, callback)
}

type WebDAVFileSystem struct {
	*gowebdav.Client
}

func NewWebDavFileSystem(host, username, password string) (*WebDAVFileSystem, error) {
	client := gowebdav.NewClient(host, username, password)
	err := client.Connect()
	return &WebDAVFileSystem{client}, err
}

func (fs *WebDAVFileSystem) WalkDir(dirPath string, callback func(parent string, info os.FileInfo) (bool, error)) error {
	return walkDirInternal(func(s string) ([]os.FileInfo, error) {
		return fs.ReadDir(s)
	}, dirPath, callback)
}

func walkDirInternal(fnReadDir func(string) ([]os.FileInfo, error), dirPath string, callback func(parent string, info os.FileInfo) (bool, error)) error {
	files, err := fnReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, fi := range files {
		if fi.IsDir() {
			err := walkDirInternal(fnReadDir, filepath.Join(dirPath, fi.Name()), callback)
			if err != nil {
				return err
			}
		}
		stop, callbackErr := callback(dirPath, fi)
		if stop || callbackErr != nil {
			return callbackErr
		}
	}
	return nil
}
