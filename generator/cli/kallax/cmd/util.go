package cmd

import "os"

func isDirectory(name string) (bool, error) {
	info, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}
