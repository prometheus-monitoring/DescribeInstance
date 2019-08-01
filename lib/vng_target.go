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
	NICS         []struct {
		Card string `json:"card"`
		MAC  string `json:"mac_address"`
		Nets []struct {
			VlanID   string `json:"vlan_id"`
			VlanType string `json:"vlan_type"`
			IP       string `json:"ip_address"`
		} `json:"network"`
	} `json:"NICS"`
	Location string `json:"LocationCode"`
	Data     struct {
		Owners []struct {
			Role string   `json:"role"`
			Name []string `json:"users"`
		} `json:"technical_owner"`
	} `json:"Data"`
}

func generateQuery(filter config.Filter, location string) string {
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
	// if len(filter.Match.SELv1) != 0 {
	// 	for _, se := range filter.Match.SELv1 {
	// 		cond := fmt.Sprintf(`Data like '%%%s%%'`, se)
	// 		conditions = append(conditions, cond)
	// 	}
	// }
	if len(filter.Match.IP) != 0 {
		for _, ip := range filter.Match.IP {
			cond := fmt.Sprintf(`NICS like '%%"%s"%%'`, ip)
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
	// if len(filter.NotMatch.SELv1) != 0 {
	// 	for _, se := range filter.NotMatch.SELv1 {
	// 		cond := fmt.Sprintf(`Data not like '%%%s%%'`, se)
	// 		conditions = append(conditions, cond)
	// 	}
	// }
	if len(filter.NotMatch.IP) != 0 {
		for _, ip := range filter.NotMatch.IP {
			cond := fmt.Sprintf(`NICS not like '%%"%s"%%'`, ip)
			conditions = append(conditions, cond)
		}
	}
	return fmt.Sprintf(`Select VMServerName, ProductAlias, LocationCode, NICS, Data from allserverinfo where LocationCode="%s" and %s`, location, strings.Join(conditions, " and "))
}

func queryData(db *sql.DB, locationCode string, filter config.Filter) (*sql.Rows, error) {
	queryStatement := generateQuery(filter, locationCode)
	results, err := db.Query(queryStatement)
	return results, err
}

func seIsExist(listSE interface{}, SE interface{}) bool {
	ls := reflect.ValueOf(listSE)
	for i := 0; i < ls.Len(); i++ {
		if reflect.TypeOf(SE).Kind().String() != "slice" {
			if strings.Contains(SE.(string), ls.Index(i).String()) {
				fmt.Println("$$$$$$$$$$$$$$$$$")
				return true
			}
		} else {
			se := reflect.ValueOf(SE)
			for j := 0; j < se.Len(); j++ {
				if strings.Contains(se.Index(j).String(), ls.Index(i).String()) {
					return true
				}
			}
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
	t.Addrs = append(t.Addrs, addr)
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
	results, err := queryData(db, locationCode, filter)
	if err != nil {
		return ts, err
	}

	logLevel.Info("[vng] Create list targets")
	for results.Next() {
		var data Data
		var nics, d string
		err = results.Scan(&data.VMServerName, &data.ProductCode, &data.Location, &nics, &d)
		if err != nil {
			logLevel.Error(err.Error())
			continue
		}
		// Unmarshal Nics
		err = json.Unmarshal([]byte(nics), &data.NICS)
		if err != nil {
			logLevel.Error(err.Error())
			continue
		}
		//Unmarshal Owner in Data json
		err = json.Unmarshal([]byte(strings.ReplaceAll(d, "\n", "\\n")), &data.Data)
		if err != nil {
			logLevel.Error(err.Error())
			continue
		}
		for _, owner := range data.Data.Owners {
			t := new(Target)
			if strings.Contains(owner.Role, "Level 1") {
				if len(filter.NotMatch.SELv1) != 0 {
					if seIsExist(filter.NotMatch.SELv1, owner.Name) {
						continue
					}
					ts = append(ts, t.append(data))
				} else if len(filter.Match.SELv1) != 0 {
					if seIsExist(filter.Match.SELv1, owner.Name) {
						ts = append(ts, t.append(data))
					}
				} else {
					ts = append(ts, t.append(data))
				}
			}
		}
	}
	return ts, err
}
