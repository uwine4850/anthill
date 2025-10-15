package pathutils

import "os"

func Exists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return os.ErrNotExist
	}
	return err
}

func IsFile(path string) (bool, error) {
	f, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !f.IsDir() {
		return true, nil
	}
	return false, nil
}
