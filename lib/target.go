package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

type Target struct {
	Addrs  []string `json:"targets"`
	Labels LabelSet `json:"labels"`
}

type LabelSet map[string]string

type Targets []Target

func writeFile(content []byte, dir string) error {
	err := ioutil.WriteFile(dir, content, 0644)
	return err
}

func (ts Targets) AddManual(dc string, logLevel *logrus.Logger) []Target {
	logLevel.Info("Add targets manual")
	if dc == "all" || dc == "" {
		fmt.Println("Please choose datacenter to add targets (aws, gcp, vng)")
	} else {
		for {
			t := Target{}
			// add address target
			fmt.Println("**Add new target: ")
			var addr string
			for {
				fmt.Print("***Address: ")
				fmt.Scan(&addr)
				i := strings.Split(addr, ":")
				//check missing port
				if len(i) != 2 {
					fmt.Println("****Your address missing port. Please try again!!!")
					continue
				}
				ip := net.ParseIP(i[0])
				if ip.To4() != nil {
					t.Addrs = append(t.Addrs, addr)
					break
				}
				fmt.Printf("****%s is not an Ipv4 address. Please try again!!!\n", ip)
			}
			//add labels
			fmt.Println("**Add labels (press key with \"-n\" stop to add labels): ")
			t.Labels = make(map[string]string)
			for {
				fmt.Print("***Key: ")
				var key string
				fmt.Scan(&key)
				if key == "-n" {
					break
				}
				fmt.Print("***Value: ")
				var value string
				fmt.Scan(&value)
				t.Labels[key] = value
			}

			fmt.Print("**Do you want to continue add new target? [y|N]")
			var cont string
			fmt.Scan(&cont)
			ts = append(ts, t)
			if strings.ToLower(cont) == "n" {
				break
			}
		}
	}
	return ts
}

func (ts Targets) NewTargetsManual(datacenter, desDir string, logLevel *logrus.Logger) {
	targets := ts.AddManual(datacenter, logLevel)
	fileDir := fmt.Sprintf("%stargets_manual_%s.json", desDir, datacenter)
	// fmt.Printf("** Select the path to save [%s]: ", fileDir)
	// var path string
	// fmt.Scan(&path)
	// if path != "" {
	// 	fileDir = path
	// }
	logLevel.Infof("Write targets to %s", fileDir)
	content, _ := json.MarshalIndent(targets, "", "\t")
	err := writeFile(content, fileDir)
	if err != nil {
		logLevel.Error(err)
	} else {
		fmt.Printf("**Write targets to %s\n", fileDir)
		logLevel.Info("Write targets manual completed")
	}
}
