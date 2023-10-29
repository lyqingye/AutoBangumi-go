package bot

import "os"

type FileSystem interface {
	ReadDir(dirPath string) ([]os.FileInfo, error)
	WalkDir(dirPath string, callback func(parent string, info os.FileInfo) (bool, error)) error
}
