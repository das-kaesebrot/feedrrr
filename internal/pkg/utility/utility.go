package utility

import (
	"log/slog"
	"net/url"
	"os"
)

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func HandleErr(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}

func CheckFileAccess(path string) error {
	// doesnt't exist -> not read/writeable
	if _, err := os.Stat(path); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
