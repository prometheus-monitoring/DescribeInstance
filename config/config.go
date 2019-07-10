package config

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type filter struct {
	Match struct {
		Status []string `yaml:"status"`
	} `yaml:"match"`
	NotMatch struct {
		Prod  []string `yaml:"product"`
		SELv1 []string `yaml:"selv1"`
	} `yaml:"not_match"`
}

type credentials struct {
	AWS   string `yaml:"aws"`
	GCP   string `yaml:"gcp"`
	MySQL mysql  `yaml:"mysql"`
}

type mysql struct {
	DBname     string `yaml:"name"`
	RemoteHost string `yaml:"remote_host"`
	User       string `yaml:"user"`
	Pass       string `yaml:"password"`
}

type Config struct {
	Creds  credentials `yaml:"credentials"`
	Filter filter      `yaml:"filter"`
}

func (conf *Config) NewConfig() {
	dir, _ := os.Getwd()
	data, err := ioutil.ReadFile(path.Join(dir, "config.yml"))
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal((data), &conf)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
}
