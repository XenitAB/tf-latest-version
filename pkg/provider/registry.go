package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type versionRoot struct {
	Version string `json:"version"`
}

func getLatestVersion(name string) (string, error) {
	if name == "" {
		return "", errors.New("name cannot be empty")
	}

	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s", name)
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	vr := &versionRoot{}
	err = json.NewDecoder(r.Body).Decode(vr)
	if err != nil {
		return "", err
	}
	if vr.Version == "" {
		return "", fmt.Errorf("latest version for %q cannot be empty", name)
	}
	return vr.Version, nil
}
