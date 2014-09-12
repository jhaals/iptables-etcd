package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Rules struct {
	Key   string
	Value string
}
type Response struct {
	Node struct {
		Nodes []Rules
	}
}

func getRules() ([]Rules, error) {
	etcdHost := os.Getenv("ETCD_HOST")
	if etcdHost == "" {
		log.Fatal("Environment variable ETCD_HOST must be set.")
	}

	resp, err := http.Get(etcdHost + "/v2/keys/iptables?sorted=true")
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rules Response
	json.Unmarshal(body, &rules)
	return rules.Node.Nodes, nil
}

func writeRules(rules []Rules, file string) error {
	// We shouldn't write less than 3 rules.
	if len(rules) < 3 {
		return errors.New("Dude! There's not enough rules in your iptables...")
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, rule := range rules {
		f.Write([]byte(rule.Value + "\n"))
	}
	return nil
}

func main() {

	rules, err := getRules()
	if err != nil {
		// No rules no run
		log.Fatal(err)
	}

	err = writeRules(rules, "/dst/iptables-rules.txt")
	if err != nil {
		log.Fatal(err)
	}
}
