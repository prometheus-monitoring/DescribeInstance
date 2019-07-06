package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/prometheus-monitoring/DescribeInstance/lib"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func writeFile(content []byte, dir string) error {
	err := ioutil.WriteFile(dir, content, 0644)
	return err
}

func ensureDir(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	return err
}

func main() {
	// New logger
	logrus.SetFormatter(&logrus.TextFormatter{})
	var loglevel = logrus.New()
	loglevel.Out = os.Stdout

	//Parse flag
	app := kingpin.New(filepath.Base(os.Args[0]), "Script get describe instance from cloud server")
	app.HelpFlag.Short('h')

	getFromAllCmd := app.Command("all", "Get describe instance from all aws, gcp, vng")
	getFromAWSCmd := app.Command("aws", "Get describe instance from aws")
	getFromGCPCmd := app.Command("gcp", "Get describe instance from gcp")
	getFromVNGCmd := app.Command("vng", "Get describe instance from vng")
	// configFiles := getFromGCPCmd.Arg(
	// 	"config-files",
	// 	"The config files to check.",
	// ).Required().ExistingFiles()
	parsedCmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	// New targets
	ts := new(lib.Targets)
	var wg sync.WaitGroup

	desDir := "targets/"
	ensureDir(desDir)

	switch parsedCmd {
	case getFromAllCmd.FullCommand():
		wg.Add(3)
		fallthrough
	case getFromAWSCmd.FullCommand():
		wg.Add(1)
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsAWS(loglevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				filedir := desDir + "targets_aws.json"
				loglevel.Info("Write all targets on aws to json file")
				err = writeFile(content, filedir)
				if err != nil {
					loglevel.Error(err)
				} else {
					loglevel.Info("Write targets on datacenter vng completed")
				}
			} else {
				loglevel.Error(err)
			}
		}()
		if !strings.Contains(parsedCmd, "all") {
			break
		}
		fallthrough
	case getFromGCPCmd.FullCommand():
		wg.Add(1)
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsGCP(loglevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				filedir := desDir + "targets_gcp.json"
				loglevel.Info("Write all targets on gcp to json file")
				err = writeFile(content, filedir)
				if err != nil {
					loglevel.Error(err)
				} else {
					loglevel.Info("Write targets on datacenter vng completed")
				}
			} else {
				loglevel.Error(err)
			}
		}()
		if !strings.Contains(parsedCmd, "all") {
			break
		}
		fallthrough
	case getFromVNGCmd.FullCommand():
		wg.Add(1)
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsVNG(loglevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				filedir := desDir + "targets_vng.json"
				loglevel.Info("Write all targets on datacenter vng to json file")
				err = writeFile(content, filedir)
				if err != nil {
					loglevel.Error(err)
				} else {
					loglevel.Info("Write targets on datacenter vng completed")
				}
			} else {
				loglevel.Error(err)
			}
		}()
	}
	wg.Wait()
}
