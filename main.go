package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func getRules(url string) (Iptables, error) {
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

func main() {
	host := "172.17.42.1:4001"
	dst := os.Getenv("DST")
	if dst == "" {
		log.Println("Destination path 'DST' must be set")
		os.Exit(1)
	}
	id := os.Getenv("ID")
	if dst == "" {
		log.Println("ID must be set")
		os.Exit(1)
	}
	rulesFile := filepath.Join(dst, "iptables.rules")
	log.Println("Starting... Going to watch etcd for changes")
	for {
		// Wait 5 sec just in case something goes bananas.
		time.Sleep(time.Second * 5)
		url := func() string {
			if _, err := os.Stat(rulesFile); err != nil {
				return "http://" + etcd + "/v2/keys/iptables?sorted=true"
			}
			return "http://" + etcd + "/v2/keys/iptables?sorted=true&wait=true"
		}
		r, err := getRules(url())
		if err != nil {
			// No response from etcd
			log.Println(err)
			continue
		}

		// We shouldn't write less than 3 rules.
		if len(r.Rules) < 3 {
			log.Println("Dude! There's not enough rules in etcd...")
			continue
		}

		// Write rules from etcd to file
		err = writeRules(r.Rules, rulesFile)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(rulesFile, "updated")
		// Fix this..
		exec.Command(fmt.Sprintf("curl -L -XPUT http://%s/v2/keys/iptables-reload/%s -d value=reload", etcd, id)).Output()
		continue
	}
}
