package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"DescribeInstance/lib"
)

func writeFile(content []byte, dir string) {
	err := ioutil.WriteFile(dir, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ts := new(lib.Targets)
	desDir := "targets/"
	switch arg := os.Args[1]; arg {
	case "all":
		fallthrough
	case "aws":
		content, _ := json.MarshalIndent(ts.GetTargetsAWS(), "", "\t")
		filedir := desDir + "target_aws.json"
		go writeFile(content, filedir)
		if arg != "all" {
			break
		}
		fallthrough
	case "gcp":
		content, _ := json.MarshalIndent(ts.GetTargetsGCP(), "", "\t")
		filedir := desDir + "target_gcp.json"
		go writeFile(content, filedir)
		if arg != "all" {
			break
		}
		fallthrough
	case "vng":
		content, _ := json.MarshalIndent(ts.GetTargetsVNG(), "", "\t")
		filedir := desDir + "target_vng.json"
		go writeFile(content, filedir)
	}
}
