package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/prometheus-monitoring/DescribeInstance/config"
	"github.com/prometheus-monitoring/DescribeInstance/lib"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	locationsVN = [...]string{"HCM_QTSC_T1", "HCM_QTSC_T2", "Singapore"}
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
	var logLevel = logrus.New()
	logLevel.Out = os.Stdout

	// Read config
	var conf config.Config
	conf.NewConfig()

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

	desDir := "/etc/prometheus/targets/"
	ensureDir(desDir)

	switch parsedCmd {
	case getFromAllCmd.FullCommand():
		wg.Add(3)
		fallthrough
	case getFromAWSCmd.FullCommand():
		if !strings.Contains(parsedCmd, "all") {
			wg.Add(1)
		}
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsAWS(logLevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				fileDir := desDir + "targets_aws.json"
				logLevel.Info("Write all targets on aws to json file")
				err = writeFile(content, fileDir)
				if err != nil {
					logLevel.Error(err)
				} else {
					logLevel.Info("Write targets on datacenter vng completed")
				}
			} else {
				logLevel.Error(err)
			}
		}()
		if !strings.Contains(parsedCmd, "all") {
			break
		}
		fallthrough
	case getFromGCPCmd.FullCommand():
		if !strings.Contains(parsedCmd, "all") {
			wg.Add(1)
		}
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsGCP(logLevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				fileDir := desDir + "targets_gcp.json"
				logLevel.Info("Write all targets on gcp to json file")
				err = writeFile(content, fileDir)
				if err != nil {
					logLevel.Error(err)
				} else {
					logLevel.Info("Write targets on datacenter vng completed")
				}
			} else {
				logLevel.Error(err)
			}
		}()
		if !strings.Contains(parsedCmd, "all") {
			break
		}
		fallthrough
	case getFromVNGCmd.FullCommand():
		// if !strings.Contains(parsedCmd, "all") {
		// 	wg.Add(1)
		// }
		go func() {
			defer wg.Done()
			for _, location := range locationsVN {
				targets, err := ts.GetTargetsVNG(logLevel, location, conf.Filter)
				if err == nil {
					content, _ := json.MarshalIndent(targets, "", "\t")
					var fileDir string
					if strings.Contains(location, "T1") {
						fileDir = fmt.Sprintf("%stargets_vng_%s.json", desDir, "oldfarm")
					} else if strings.Contains(location, "T2") {
						fileDir = fmt.Sprintf("%stargets_vng_%s.json", desDir, "newfarm")
					} else {
						fileDir = fmt.Sprintf("%stargets_vng_%s.json", desDir, "singapore")
					}
					logLevel.Info("Write all targets on datacenter vng to json file")
					err = writeFile(content, fileDir)
					if err != nil {
						logLevel.Error(err)
					} else {
						logLevel.Info("Write targets on datacenter vng completed")
					}
				} else {
					logLevel.Error(err)
				}
			}
		}()
	}
	wg.Wait()
}
