package lib

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type Data struct {
	VMServerName string `json:"VMServerName"`
	ProductCode  string `json:"ProductAlias"`
	NICS         []nic  `json:"NICS"`
	Location     string `json:"LocationCode"`
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

func querydata(db *sql.DB, locationCode string) (*sql.Rows, error) {
	queryStatement := fmt.Sprintf(`Select VMServerName, ProductAlias, LocationCode, NICS from allserverinfo
																		where NOT Data like '%s'
																			and NOT Data like '%s'
																			and	NOT Data like '%s'
																			and LocationCode='%s'`, "%truongln%", "%phongnvd%", "%vihct%", locationCode)
	results, err := db.Query(queryStatement)
	return results, err
}

func (t *Target) append(d Data) Target {
	t.Labels = make(map[string]string)
	t.Labels["instance"] = d.VMServerName
	t.Labels["product_code"] = d.ProductCode
	for _, nic := range d.NICS {
		for _, net := range nic.Nets {
			if net.VlanType == "public" && (t.Labels["ip"] == "" || t.Labels["ip_priv"] == "") {
				t.Labels["ip"] = net.IP
			} else {
				t.Labels["ip_priv"] = net.IP
			}
		}
	}
	t.Labels["location_code"] = "vng_" + d.Location
	addr := t.Labels["ip_priv"] + ":11011"

	t.Targets = append(t.Targets, addr)
	return *t
}

func (ts Targets) GetTargetsVNG(logLevel *logrus.Logger, locationCode string) ([]Target, error) {
	logLevel.Info("[vng] Establishing connection to database")
	db, err := connect()
	defer db.Close()
	if err != nil {
		return ts, err
	}
	logLevel.Info("[vng] Ensure the connection is established")
	err = db.Ping()
	if err != nil {
		return ts, err
	}

	logLevel.Info("[vng] Query data")
	results, err := querydata(db, locationCode)
	if err != nil {
		return ts, err
	}

	logLevel.Info("[vng] Create list targets")
	for results.Next() {
		var data Data
		var nics string
		err = results.Scan(&data.VMServerName, &data.ProductCode, &data.Location, &nics)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal([]byte(nics), &data.NICS)
		if err != nil {
			panic(err.Error())
		}
		t := new(Target)
		ts = append(ts, t.append(data))
	}
	return ts, err
}
