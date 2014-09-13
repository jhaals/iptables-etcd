package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
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
	url := "http://172.17.42.1:4001/v2/keys/iptables?sorted=true"
	resp, err := http.Get(url)
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

func Checksum(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

func main() {
	tempFile := "/iptables.rules"
	destinationFile := "/dst/iptables.rules"
	for {
		rules, err := getRules()
		if err != nil {
			// No rules no run
			log.Fatal(err)
		}
		// We shouldn't write less than 3 rules.
		if len(rules) < 3 {
			log.Fatal("Dude! There's not enough rules in your iptables...")
		}

		// Write rules from etcd to file
		err = writeRules(rules, tempFile)
		if err != nil {
			log.Fatal(err)
		}

		tempFileChecksum, _ := Checksum(tempFile)
		destinationFileChecksum, _ := Checksum(destinationFile)
		//Checksum tempfile with existing file. Only overwrite if necessary
		if tempFileChecksum != destinationFileChecksum {
			err = writeRules(rules, destinationFile)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Rules updated")
			os.Exit(0)
		}
		time.Sleep(time.Second * 30)
	}
}
