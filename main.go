package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/prometheus-monitoring/DescribeInstance/lib"
)

func writeFile(content []byte, dir string) {
	err := ioutil.WriteFile(dir, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func ensureDir(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	ts := new(lib.Targets)

	var wg sync.WaitGroup

	desDir := "targets/"
	ensureDir(desDir)

	switch arg := os.Args[1]; arg {
	case "all":
		wg.Add(3)
		fallthrough
	case "aws":
		wg.Add(1)
		content, _ := json.MarshalIndent(ts.GetTargetsAWS(), "", "\t")
		filedir := desDir + "target_aws.json"
		go func() {
			defer wg.Done()
			writeFile(content, filedir)
		}()
		if arg != "all" {
			break
		}
		fallthrough
	case "gcp":
		wg.Add(1)
		content, _ := json.MarshalIndent(ts.GetTargetsGCP(), "", "\t")
		filedir := desDir + "target_gcp.json"
		go func() {
			defer wg.Done()
			writeFile(content, filedir)
		}()
		if arg != "all" {
			break
		}
		fallthrough
	case "vng":
		wg.Add(1)
		content, _ := json.MarshalIndent(ts.GetTargetsVNG(), "", "\t")
		filedir := desDir + "target_vng.json"
		go func() {
			defer wg.Done()
			writeFile(content, filedir)
		}()
	}
	wg.Wait()
}
