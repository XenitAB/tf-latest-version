package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Registry interface {
	getLatestVersion(name string) (string, error)
}

type HashicorpRegistry struct {
	cache map[string]string
}

func NewHashicorpRegistry() HashicorpRegistry {
	return HashicorpRegistry{
		cache: map[string]string{},
	}
}

type versionRoot struct {
	Version string `json:"version"`
}

func (h HashicorpRegistry) getLatestVersion(name string) (string, error) {
	if name == "" {
		return "", errors.New("name cannot be empty")
	}

	// no need to lookup if latest version is cached
	if version, ok := h.cache[name]; ok {
		return version, nil
	}

	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s", name)
	client := http.Client{Timeout: 10 * time.Second}
	r, err := client.Get(url)
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

	v := vr.Version
	h.cache[name] = v
	return v, nil
}

type FakeRegistry struct {
	providers map[string][]string
}

func (f FakeRegistry) getLatestVersion(name string) (string, error) {
	versions, ok := f.providers[name]
	if !ok {
		return "", fmt.Errorf("provider %q not found", name)
	}

	return versions[0], nil
}
