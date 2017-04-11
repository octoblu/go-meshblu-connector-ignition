package forever

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/inconshreveable/go-update"
	"github.com/kardianos/osext"
)

// VersionInfo defines the information of the request
type VersionInfo struct {
	Version string `json:"version"`
}

func shouldUpdate(currentVersion, latestVersion string) bool {
	return fmt.Sprintf("v%s", currentVersion) != latestVersion
}

func doUpdate(version string) error {
	downloadURL := getDownloadURL(version)
	res, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Unexpected response status code %v", res.StatusCode)
	}
	defer res.Body.Close()
	err = update.Apply(res.Body, update.Options{})
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			return rerr
		}
		return err
	}
	return nil
}

func getDownloadURL(version string) string {
	baseURL := "https://github.com/octoblu/go-meshblu-connector-ignition/releases/download"
	return fmt.Sprintf("%s/%s/meshblu-connector-ignition-%s-%s", baseURL, version, runtime.GOOS, runtime.GOARCH)
}

func resolveLatestVersion() (string, error) {
	var versionInfo VersionInfo
	url := "https://connector-service.octoblu.com/releases/octoblu/go-meshblu-connector-ignition/latest/version/resolve"
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("Unexpected response status code %v", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if len(body) == 0 {
		return "", nil
	}
	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		return "", err
	}
	return versionInfo.Version, nil
}

func startNew() error {
	ignitionScript, err := osext.Executable()
	if err != nil {
		return err
	}
	err = exec.Command(ignitionScript).Start()
	if err != nil {
		return err
	}
	return nil
}
