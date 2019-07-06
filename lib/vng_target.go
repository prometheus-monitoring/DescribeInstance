package lib

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type Data struct {
	VMServerName string `json:"VMServerName"`
	ProductCode  string `json:"ProductAlias"`
	NICS         []nic  `json:"NICS"`
	// Owners       []se   `json:"technical_owner"`
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

// type se struct {
// 	Role  string   `json:"role"`
// 	Users []string `json:"users"`
// }

func connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", "discoverylocal:1qaz8ik,@tcp(61.28.251.123)/getlistserver_sdk")
	return db, err
}

func (ts Targets) GetTargetsVNG(loglevel *logrus.Logger) ([]Target, error) {
	loglevel.Info("[vng] Establishing connection to database")
	db, err := connect()
	defer db.Close()
	if err != nil {
		return ts, err
	}
	loglevel.Info("[vng] Ensure the connection is established")
	err = db.Ping()
	if err != nil {
		return ts, err
	}
	loglevel.Info("[vng] Query data")
	results, err := db.Query(`Select VMServerName, ProductAlias, NICS from allserverinfo
														where NOT Data like '%truongln%'
															and NOT Data like '%phongnvd%'
															and	NOT Data like '%vihct%'`)
	if err != nil {
		return ts, err
	}

	loglevel.Info("[vng] Create list targets")
	for results.Next() {
		var data Data
		var nics string
		err = results.Scan(&data.VMServerName, &data.ProductCode, &nics)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal([]byte(nics), &data.NICS)
		if err != nil {
			panic(err.Error())
		}

		// Filter SE level 1
		// for _, owner := range data.Owners {
		// 	if owner.Role == "SE Level 1" {
		// 		for _, user := range owner.Users {
		// 			if filter(user) {
		// 				break
		// 			}
		// 		}
		// 	}
		// }

		t := new(Target)
		t.Labels = make(map[string]string)
		// t.Labels["zone"] =
		t.Labels["instance"] = data.VMServerName
		t.Labels["product_code"] = data.ProductCode
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
			if _, ok := t.Labels["ip_priv"]; ok {
				addr = t.Labels["ip_priv"] + ":11011"
			} else {
				continue
			}
		}

		t.Targets = append(t.Targets, addr)
		ts = append(ts, *t)
	}
	return ts, err
}
