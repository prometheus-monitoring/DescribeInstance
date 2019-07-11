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

type Credentials struct {
	AWS   string `yaml:"aws"`
	GCP   string `yaml:"gcp"`
	MySQL Mysql  `yaml:"mysql"`
}

type Mysql struct {
	DBname     string `yaml:"name"`
	RemoteHost string `yaml:"remote_host"`
	User       string `yaml:"user"`
	Pass       string `yaml:"password"`
}

type Config struct {
	Creds  Credentials `yaml:"credentials"`
	Filter Filter      `yaml:"filter"`
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
