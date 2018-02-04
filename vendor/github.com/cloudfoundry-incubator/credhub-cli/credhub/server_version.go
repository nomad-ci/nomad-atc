package credhub

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/credhub-cli/credhub/server"
	"github.com/hashicorp/go-version"
)

func (ch *CredHub) ServerVersion() (*version.Version, error) {
	info, err := ch.Info()
	if err != nil {
		return nil, err
	}
	v := info.App.Version
	if v == "" {
		v, err = ch.getVersion()
		if err != nil {
			return nil, err
		}
	}
	return version.NewVersion(v)
}

func (ch *CredHub) getVersion() (string, error) {
	response, err := ch.Request("GET", "/version", nil, nil)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	versionData := &server.VersionData{}
	decoder := json.NewDecoder(response.Body)

	if err = decoder.Decode(&versionData); err != nil {
		return "", err
	}

	return versionData.Version, nil
}
