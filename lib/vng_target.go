package lib

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
)

var (
	SELV1 = [...]string{"phongnvd@vng.com.vn", "vihct@vng.com.vn", "truongln@vng.com.vn"}
)

type Data struct {
	VMServerName string `json:"vm_server_name"`
	NICS         []nic  `json:"nics"`
	Owners       []se   `json:technical_owner`
}

type nic struct {
	Card string    `json:"card"`
	MAC  string    `json:"mac_address"`
	Nets []network `json:"network"`
}

type network struct {
	VlanID   string `json:"vlan_id"`
	VlanType string `json:"vlan_type"`
	IP       string `json:"ip_address"`
}

type se struct {
	Role  string   `json:"role"`
	Users []string `json:"users"`
}

func connect() *sql.DB {
	db, err := sql.Open("mysql", "discoverylocal:1qaz8ik,@tcp(127.0.0.1:3306)/getlistserver_sdk")
	if err != nil {
		panic(err.Error())
	}
	return db
}

func filter(name string) bool {
	for _, owner := range SELV1 {
		if name == owner {
			return true
		}
	}
	return false
}

func (ts Targets) GetTargetsVNG() []Target {
	db := connect()
	defer db.Close()
	err := db.Ping()
	if err != nil {
		panic(err.Error())
	}

	results, err := db.Query("SELECT Data FROM allserverinfo")
	if err != nil {
		panic(err.Error())
	}

	for results.Next() {
		var data Data
		var dataString string
		err = results.Scan(&dataString)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal([]byte(dataString), &data)
		if err != nil {
			panic(err.Error())
		}

		// Filter SE level 1
		for _, owner := range data.Owners {
			if owner.Role == "SE Level 1" {
				for _, user := range owner.Users {
					if filter(user) {
						break
					}
				}
			}
		}

		t := new(Target)
		t.Labels = make(map[string]string)
		// t.Labels["zone"] =
		t.Labels["hostname"] = data.VMServerName
		for _, nic := range data.NICS {
			for _, net := range nic.Nets {
				if net.VlanType == "public" && (t.Labels["ip"] == "" || t.Labels["ip_priv"] == "") {
					t.Labels["ip"] = net.IP
				} else {
					t.Labels["ip_priv"] = net.IP
				}
			}
		}

		addr := t.Labels["ip"] + ":11011"
		if _, ok := t.Labels["ip"]; !ok {
			addr = t.Labels["ip_priv"] + ":11011"
		}

		t.Targets = append(t.Targets, addr)
		ts = append(ts, *t)
	}
	return ts
}
