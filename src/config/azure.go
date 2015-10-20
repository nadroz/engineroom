package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

var (
	ErrEnvironmentNotFound = errors.New("undefined environment")
)

type AzureConfig struct {
	Name                  string
	AccessKey             string
	ManagementCertificate []byte
}

func LoadAzureConfig(configFile, environment string) (*AzureConfig, error) {

	type envInfo struct {
		Name                      string `toml:"storage_account_name"`
		AccessKey                 string `toml:"storage_account_access_key"`
		ManagementCertificatePath string `toml:"management_certificate"`
	}

	var config map[string]envInfo
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, err
	}

	env, ok := config[environment]
	if !ok {
		return nil, ErrEnvironmentNotFound
	}

	if env.ManagementCertificatePath != "" {
		buf, err := ioutil.ReadFile(env.ManagementCertificatePath)
		if err != nil {
			return nil, err
		}
		return &AzureConfig{env.Name, env.AccessKey, buf}, nil
	}

	return &AzureConfig{env.Name, env.AccessKey, nil}, nil
}
