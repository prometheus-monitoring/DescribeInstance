package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// type instanceInfo struct {
// Project   string `json:"project"`
// 	Zone      string `json:"zone"`
// 	Hostname  string `json:"hostname"`
// 	IPpublic  string `json:"ipPublic"`
// 	IPprivate string `json:"ipPrivate"`
// }

type target struct {
	Targets []string `json:"target"`
	Labels  LabelSet `json:"labels"`
}

type LabelSet map[string]string

func main() {
	var listTargets []target
	// gcp := GetTargets(listTargets)
	// listTargets = append(listTargets, gcp...)
	aws := GetTargetsAWS(listTargets)
	listTargets = append(listTargets, aws...)
	file, _ := json.MarshalIndent(listTargets, "", "\t")
	err := ioutil.WriteFile("target.json", file, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
