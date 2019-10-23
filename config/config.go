package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Filter struct {
	Match    filterEle `yaml:"match"`
	NotMatch filterEle `yaml:"not_match"`
}
type filterEle struct {
	Status []string `yaml:"status"`
	Prod   []string `yaml:"product"`
	SELv1  []string `yaml:"selv1"`
	IP     []string `yaml:"ip"`
}

type InfoCloud struct {
	CredentialsPath string    `yaml:"credentials_path"`
	MySQL           InfoMySQL `yaml:"mysql"`
	Filter          Filter    `yaml:"filter"`
}

type InfoMySQL struct {
	DBname     string `yaml:"name"`
	Pass       string `yaml:"password"`
	RemoteHost string `yaml:"remote_host"`
	User       string `yaml:"user"`
}

type Config struct {
	AWS InfoCloud `yaml:"aws"`
	GCP InfoCloud `yaml:"gcp"`
	VNG InfoCloud `yaml:"vng"`
}

func (conf *Config) NewConfig(logLevel *logrus.Logger, confPath string) {
	data, err := ioutil.ReadFile(confPath)
	if err != nil {
		logLevel.Fatal(err)
	}
	err = yaml.Unmarshal((data), &conf)
	if err != nil {
		logLevel.Fatalf("cannot unmarshal data: %v", err)
	}
}
