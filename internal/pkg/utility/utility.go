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
