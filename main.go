package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type target struct {
	Targets []string `json:"target"`
	Labels  LabelSet `json:"labels"`
}

type LabelSet map[string]string
type Targets []target

func main() {
	ts := new(Targets)
	var content []byte
	var filedir string

	switch arg := os.Args[1]; arg {
	case "all":
		fallthrough
	case "aws":
		content, _ = json.MarshalIndent(ts.GetTargetsAWS(), "", "\t")
		filedir = "target_aws.json"
		if arg != "all" {
			break
		}
		fallthrough
	case "gcp":
		content, _ = json.MarshalIndent(ts.GetTargetsGCP(), "", "\t")
		filedir = "target_gcp.json"
		if arg != "all" {
			break
		}
		fallthrough
	case "vng":
		content, _ = json.MarshalIndent(ts.GetTargetsVNG(), "", "\t")
		filedir = "target_vng.json"
	}
	err := ioutil.WriteFile(filedir, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
