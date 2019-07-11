package lib

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus-monitoring/DescribeInstance/config"
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

func generateQuery(filter config.Filter) string {
	conditions := []string{}
	// Match
	if len(filter.Match.Status) != 0 {
		cond := fmt.Sprintf(`Status in ("%s")`, strings.Join(filter.Match.Status, `", "`))
		conditions = append(conditions, cond)
	}
	if len(filter.Match.Prod) != 0 {
		cond := fmt.Sprintf(`ProductAlias in ("%s")`, strings.Join(filter.Match.Prod, `", "`))
		conditions = append(conditions, cond)
	}
	if len(filter.Match.SELv1) != 0 {
		for _, se := range filter.Match.SELv1 {
			cond := fmt.Sprintf(`Data like '%%%s%%'`, se)
			conditions = append(conditions, cond)
		}
	}
	if len(filter.Match.IP) != 0 {
		for _, ip := range filter.Match.IP {
			cond := fmt.Sprintf(`NICS like '%%%s%%'`, ip)
			conditions = append(conditions, cond)
		}
	}
	// Not Match
	if len(filter.NotMatch.Status) != 0 {
		cond := fmt.Sprintf(`Status not in ("%s")`, strings.Join(filter.NotMatch.Status, `", "`))
		conditions = append(conditions, cond)
	}
	if len(filter.NotMatch.Prod) != 0 {
		cond := fmt.Sprintf(`ProductAlias not in ("%s")`, strings.Join(filter.NotMatch.Prod, `", "`))
		conditions = append(conditions, cond)
	}
	if len(filter.NotMatch.SELv1) != 0 {
		for _, se := range filter.NotMatch.SELv1 {
			cond := fmt.Sprintf(`Data not like '%%%s%%'`, se)
			conditions = append(conditions, cond)
		}
	}
	if len(filter.NotMatch.IP) != 0 {
		for _, ip := range filter.NotMatch.IP {
			cond := fmt.Sprintf(`NICS not like '%%%s%%'`, ip)
			conditions = append(conditions, cond)
		}
	}
	return fmt.Sprintf(`Select VMServerName, ProductAlias, LocationCode, NICS from allserverinfo where %s`, strings.Join(conditions, " and "))
}

func querydata(db *sql.DB, locationCode string, filter config.Filter) (*sql.Rows, error) {
	queryStatement := generateQuery(filter)
	results, err := db.Query(queryStatement)
	return results, err
}

func elementExist(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)
	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}
	return false
}

func (t *Target) append(d Data) Target {
	t.Labels = make(map[string]string)
	t.Labels["instance"] = d.VMServerName
	t.Labels["product_code"] = d.ProductCode
	for _, nic := range d.NICS {
		for _, net := range nic.Nets {
			if _, ok := t.Labels["ip"]; !ok && net.VlanType == "public" {
				if net.IP != "" {
					t.Labels["ip"] = net.IP
				}
			} else if _, ok := t.Labels["ip_priv"]; !ok && net.VlanType == "private" {
				if net.IP != "" {
					t.Labels["ip_priv"] = net.IP
				}
			}
		}
	}
	addr := fmt.Sprintf("%s:%s", t.Labels["ip_priv"], "11011")
	t.Targets = append(t.Targets, addr)
	return *t
}

func (ts Targets) Connect(cred config.Mysql) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", cred.User, cred.Pass, cred.RemoteHost, cred.DBname))
	return db, err
}

func (ts Targets) GetTargetsVNG(logLevel *logrus.Logger, db *sql.DB, locationCode string, filter config.Filter) ([]Target, error) {

	logLevel.Info("[vng] Ensure the connection is established")
	err := db.Ping()
	if err != nil {
		return ts, err
	}

	logLevel.Info("[vng] Query data")
	results, err := querydata(db, locationCode, filter)
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
