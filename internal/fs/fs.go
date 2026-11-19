package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mah0x211/lenv/internal/cli"
)

func Lstat(file string) (os.FileInfo, error) {
	info, err := os.Lstat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return info, nil
}

func IsDir(file string) (bool, error) {
	if info, err := Lstat(file); err != nil {
		return false, cli.Errorf("failed to lstat():", err)
	} else if info != nil {
		return info.IsDir(), nil
	}
	return false, nil
}

func getFileMode(file string) (os.FileMode, error) {
	if info, err := Lstat(file); err != nil {
		return 0, err
	} else if info != nil {
		return info.Mode(), nil
	}
	return 0, nil
}

func Mkdir(dirname string) error {
	if ok, err := IsDir(dirname); err != nil {
		return err
	} else if ok {
		return nil
	} else if err = os.MkdirAll(dirname, 0777); err != nil {
		return err
	}
	return nil
}

func CreateSymlink(oldname, newname string) error {
	var tmpName string
	if f, err := os.CreateTemp("./", "tmp_symlink_*"); err != nil {
		cli.Fatalf("failed to tempfile(): %v", err)
	} else {
		f.Close()
		os.Remove(f.Name())
		tmpName = f.Name()
	}

	if err := os.Symlink(oldname, tmpName); err != nil {
		return err
	}

	if dir := filepath.Dir(newname); len(dir) > 0 {
		if err := Mkdir(dir); err != nil {
			return err
		}
	}

	return os.Rename(tmpName, newname)
}

func ReadSymlink(file string) (string, error) {
	if mode, err := getFileMode(file); err != nil {
		return "", err
	} else if mode == 0 {
		return "", nil
	} else if mode&os.ModeSymlink == 0 {
		return "", ErrNotSymlink
	} else if pathname, err := os.Readlink(file); err != nil {
		return "", err
	} else {
		return pathname, nil
	}
}

func createFile(file string, perm os.FileMode) (*os.File, error) {
	if err := Mkdir(filepath.Dir(file)); err != nil {
		return nil, err
	}

	f, err := os.Create(file)
	if err != nil {
		return nil, err
	} else if perm > 0 {
		if err = f.Chmod(perm); err != nil {
			os.Remove(file)
			return nil, err
		}
	}

	return f, nil
}

func OpenFile(file string) (*os.File, error) {
	f, err := os.Open(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	return f, nil
}

func WriteFile(file string, perm os.FileMode, src io.Reader) error {
	f, err := createFile(file, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, src); err != nil {
		os.Remove(file)
		return err
	} else if err = f.Sync(); err != nil {
		os.Remove(file)
		return err
	}

	return nil
}

var ErrNotSymlink = fmt.Errorf("not a symlink")

func EvalSymlink(file string) (string, error) {
	file = filepath.Clean(file)
	if mode, err := getFileMode(file); err != nil {
		return "", err
	} else if mode == 0 {
		return "", nil
	} else if mode&os.ModeSymlink == 0 {
		return "", ErrNotSymlink
	} else if pathname, err := filepath.EvalSymlinks(file); err != nil {
		return "", err
	} else {
		return pathname, nil
	}
}
