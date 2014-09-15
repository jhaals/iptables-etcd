package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Key   string
	Value string
}
type Result struct {
	Entries  []Entry `json:"nodes"`
	Modified int     `json:"ModifiedIndex"`
	Created  int     `json:"CreatedIndex"`
}
type Response struct {
	Result Result `json:"node"`
}

func etcdGet(url string) (Result, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var response Response
	if err != nil {
		return response.Result, err
	}
	json.Unmarshal(body, &response)
	return response.Result, nil
}

func writeRules(rules []Entry, file string) error {
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

func UpdateRules(host string, file string) error {
	r, err := etcdGet("http://" + host + "/v2/keys/iptables/rules?sorted=true")
	if err != nil {
		// No response from etcd
		return err
	}

	// We shouldn't write less than 3 rules.
	if len(r.Entries) < 3 {
		return errors.New("Dude! There's not enough rules in etcd...")
	}

	// Write rules from etcd to file
	err = writeRules(r.Entries, file)
	if err != nil {
		return err
	}
	log.Println(file, "updated")
	return nil
}

func EtcdWatch(host string) {
	log.Println("Watching etcd for changes")
	for {
		time.Sleep(time.Second)
		_, err := etcdGet("http://" + host + "/v2/keys/iptables/reload?wait=true")
		if err != nil {
			log.Println(err)
			continue
		}
		return
	}
}
func main() {
	host := "172.17.42.1:4001"
	dst := os.Getenv("DST")
	if dst == "" {
		log.Println("Destination path 'DST' must be set")
		os.Exit(1)
	}
	for {
		EtcdWatch(host)
		err := UpdateRules(host, filepath.Join(dst, "iptables.rules"))
		if err != nil {
			log.Println(err)
		}
	}
}
