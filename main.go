package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Rule struct {
	Key   string
	Value string
}
type Iptables struct {
	Rules    []Rule `json:"nodes"`
	Modified int    `json:"ModifiedIndex"`
	Created  int    `json:"CreatedIndex"`
}
type Response struct {
	Iptables Iptables `json:"node"`
}

func getRules() (Iptables, error) {
	url := "http://172.17.42.1:4001/v2/keys/iptables?sorted=true"
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var response Response
	if err != nil {
		return response.Iptables, err
	}
	json.Unmarshal(body, &response)
	return response.Iptables, nil
}

func writeRules(rules []Rule, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write([]byte("*filter\n"))
	for _, rule := range rules {
		f.Write([]byte(rule.Value + "\n"))
	}
	f.Write([]byte("COMMIT\n"))
	return nil
}

func IptablesUpToDate(file string, index string) bool {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	if string(data) == index {
		return true
	}
	return false
}

func main() {
	dst := os.Getenv("DST")
	if dst == "" {
		log.Println("Destination path 'DST' must be set")
		os.Exit(1)
	}
	rulesFile := filepath.Join(dst, "iptables.rules")
	indexFile := filepath.Join(dst, "iptables.index")
	log.Println("Starting... Will check for updates in etcd with a 5 second interval")
	for {
		time.Sleep(time.Second * 5)
		r, err := getRules()
		if err != nil {
			// No response from etcd
			log.Println(err)
			continue
		}
		modified := strconv.Itoa(r.Modified)
		if IptablesUpToDate(indexFile, modified) {
			continue
		}

		// We shouldn't write less than 3 rules.
		if len(r.Rules) < 3 {
			log.Println("Dude! There's not enough rules in your iptables...")
			continue
		}

		// Write rules from etcd to file
		err = writeRules(r.Rules, rulesFile)
		if err != nil {
			log.Println(err)
			continue
		}
		ioutil.WriteFile(indexFile, []byte(modified), 0644)
		log.Println(rulesFile, "updated")
		return
	}
}
