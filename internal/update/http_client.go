package update

import (
	"net/http"
	"os"
	"time"
)

const userAgent = "muxdev-updater"

var httpClient = &http.Client{Timeout: 30 * time.Second}

func setRequestHeaders(req *http.Request, versionShort string) {
	req.Header.Set("User-Agent", userAgent+"/"+versionShort)
	applyAuth(req)
}

func applyAuth(req *http.Request) {
	token := os.Getenv("MUXDEV_UPDATE_TOKEN")
	if token == "" {
		return
	}
	if user := os.Getenv("MUXDEV_UPDATE_USER"); user != "" {
		req.SetBasicAuth(user, token)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
}
