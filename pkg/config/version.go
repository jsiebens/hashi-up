package config

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"io/ioutil"
	"net/http"
	"time"
)

type Versions struct {
	Nomad  string `hcl:"nomad"`
	Consul string `hcl:"consul"`
	Vault  string `hcl:"vault"`
}

func GetVersion() (*Versions, error) {
	url := "https://hashi-up.dev/versions.hcl"

	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response code %d", res.StatusCode)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result := Versions{}

	err = hclsimple.Decode("versions.hcl", body, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
