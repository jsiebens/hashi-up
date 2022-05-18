package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/Masterminds/semver"
)

type Product struct {
}

type Version struct {
	Version string `json:"version"`
}

func GetLatestVersion(product string) (string, error) {
	url := fmt.Sprintf("https://api.releases.hashicorp.com/v1/releases/%s?license_class=oss", product)

	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("invalid response code %d", res.StatusCode)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var result []Version
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	for _, i := range result {
		v, err := semver.NewVersion(i.Version)
		if err == nil && len(v.Metadata()) == 0 && len(v.Prerelease()) == 0 {
			return i.Version, nil
		}
	}

	return "", fmt.Errorf("unable to find latest version of %s", product)
}

func GetDownloadURL(product, arch string, version *semver.Version) string {
	v := semver.MustParse("1.10.4")

	if arch == "arm" && product == "consul" && version.LessThan(v) {
		arch = "armhfv6"
	}

	return fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s_%s_%s_%s.zip", product, version, product, version, runtime.GOOS, arch)
}

func GetArmSuffix(product string, version string) string {
	if product == "consul" {
		m := semver.MustParse("1.10.4")
		v := semver.MustParse(version)
		if v.LessThan(m) {
			return "armhfv6"
		}
	}
	return "arm"
}
