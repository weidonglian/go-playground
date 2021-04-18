package md5

import (
	"context"
	"crypto/md5"
	"io/fs"
	"io/ioutil"
	"path/filepath"
)

func Md5AllSeq(ctx context.Context, root string) (map[string]Md5Sum, error) {
	m := make(map[string]Md5Sum)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() == context.Canceled {
			return context.Canceled
		}

		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		m[path] = md5.Sum(data)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return m, nil
}
