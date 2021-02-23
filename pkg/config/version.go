package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sort"
	"time"

	"github.com/Masterminds/semver"
)

type Versions struct {
	Name     string                 `json:"vault"`
	Versions map[string]interface{} `json:"versions"`
}

func GetLatestVersion(product string) (string, error) {
	url := fmt.Sprintf("https://releases.hashicorp.com/%s/index.json", product)

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
	result := Versions{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	vs := make([]*semver.Version, 0)
	for i := range result.Versions {
		v, err := semver.NewVersion(i)
		if err == nil && len(v.Metadata()) == 0 && len(v.Prerelease()) == 0 {
			vs = append(vs, v)
		}
	}

	if len(vs) == 0 {
		return "", fmt.Errorf("unable to find default version of %s", product)
	}

	sort.Sort(sort.Reverse(semver.Collection(vs)))

	return vs[0].String(), nil
}

func GetDownloadURL(product, version string) string {
	var arch = runtime.GOARCH

	if arch == "arm" && product == "consul" {
		arch = "armhfv6"
	}

	return fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s_%s_%s_%s.zip", product, version, product, version, runtime.GOOS, arch)
}
