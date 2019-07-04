package main

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
)

type nic struct {
	Card string    `json:"card"`
	MAC  string    `json:"mac_address"`
	Nets []network `json:"network"`
}

type Info struct {
	VMServerName string `json:"VMServerName"`
	NICS         []nic  `json:"NICS"`
}

type network struct {
	VlanID   string `json:"vlan_id"`
	VlanType string `json:"vlan_type"`
	IP       string `json:"ip_address"`
}

func connect() *sql.DB {
	db, err := sql.Open("mysql", "discoverylocal:1qaz8ik,@tcp(127.0.0.1:3306)/getlistserver_sdk")
	if err != nil {
		panic(err.Error())
	}
	return db
}

func GetTargetsVNG(list []target) []target {
	db := connect()
	defer db.Close()
	err := db.Ping()
	if err != nil {
		panic(err.Error())
	}

	results, err := db.Query("SELECT VMServerName, NICS  FROM allserverinfo")
	if err != nil {
		panic(err.Error())
	}

	for results.Next() {
		var info Info
		var nics string
		err = results.Scan(&info.VMServerName, &nics)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal([]byte(nics), &info.NICS)
		if err != nil {
			panic(err.Error())
		}
		t := new(target)
		t.Labels = make(map[string]string)
		// t.Labels["zone"] =
		t.Labels["hostname"] = info.VMServerName
		for _, nic := range info.NICS {
			for _, net := range nic.Nets {
				if net.VlanType == "public" && (t.Labels["ip"] == "" || t.Labels["ip_priv"] == "") {
					t.Labels["ip"] = net.IP
				} else {
					t.Labels["ip_priv"] = net.IP
				}
			}
		}
		addr := t.Labels["ip"] + ":11011"
		t.Targets = append(t.Targets, addr)
		list = append(list, *t)
	}
	return list
}
