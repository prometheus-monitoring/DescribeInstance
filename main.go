package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/prometheus-monitoring/DescribeInstance/config"
	"github.com/prometheus-monitoring/DescribeInstance/lib"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	locationsVN = [...]string{"HCM_QTSC_T1", "HCM_QTSC_T2", "Singapore"}
	configPath  = kingpin.Flag("config.file", "DescribeInstance configuration file path.").Short(rune('c')).Default("config.yml").String()
	manual      = kingpin.Flag("add.manual", "Add targets munual").Short(rune('m')).Default("false").Bool()
	datacenter  = kingpin.Flag("datacenter", "Choose data center:\n\t all: Get all targets from the data center include aws, gcp, vng\n\t aws: Get all targets from the amazone web services\n\t gcp: Get all targets from the google cloud\n\t vng: Get all targets from the VN data center").Short(rune('d')).String()
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

	//Parse flag
	kingpin.HelpFlag.Short(rune('h'))
	kingpin.Parse()
	// Read config
	var conf config.Config
	conf.NewConfig(logLevel, *configPath)

	// New targets
	ts := new(lib.Targets)
	var wg sync.WaitGroup

	desDir := "/etc/prometheus/targets/"
	ensureDir(desDir)
	if *manual {
		ts.AddManual()
	} else {
		switch *datacenter {
		case "all":
			wg.Add(3)
			fallthrough
		case "aws":
			if *datacenter != "all" {
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
			if *datacenter == "all" {
				break
			}
			fallthrough
		case "gcp":
			if *datacenter != "all" {
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
			if *datacenter == "all" {
				break
			}
			fallthrough
		case "vng":
			if *datacenter != "all" {
				wg.Add(1)
			}
			go func() {
				defer wg.Done()
				logLevel.Info("[vng] Establishing connection to database")
				db, err := ts.Connect(conf.Creds.MySQL)
				defer db.Close()
				if err != nil {
					logLevel.Panic(err)
				}
				for _, location := range locationsVN {
					targets, err := ts.GetTargetsVNG(logLevel, db, location, conf.Filter)
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
	}
	wg.Wait()
}
