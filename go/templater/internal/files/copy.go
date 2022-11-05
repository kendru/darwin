package files

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
)

type transformFunc = func(data []byte) ([]byte, error)

func CopyWithTransform(srcFile, dstFile string, perm fs.FileMode, xf transformFunc) error {
	in, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}

	out, err := xf(in)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dstFile, out, perm)
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	defer func() {
		closeErr := in.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}
