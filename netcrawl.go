package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/johnsiilver/netcrawl/explorer"
	"github.com/johnsiilver/netcrawl/explorer/config"
)

var (
	rootNode = flag.String("root", "", "The IP/Hostname of the root device")
)

func exitf(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Println()
	os.Exit(1)
}

func loadConfig() (config.Config, error) {
	const withBinary = "./netcrawl.conf"
	const inETC = "/etc/netcrawl.conf"

	conf := config.Config{}
	if _, err := os.Stat(withBinary); err == nil {
		b, err := ioutil.ReadFile(withBinary)
		if err != nil {
			return conf, err
		}
		if err := json.Unmarshal(b, &conf); err != nil {
			return conf, err
		}
	}
	if _, err := os.Stat(inETC); err == nil {
		b, err := ioutil.ReadFile(inETC)
		if err != nil {
			return conf, err
		}
		if err := json.Unmarshal(b, &conf); err != nil {
			return conf, err
		}
	}

	return conf, fmt.Errorf("netcrawl.conf not found in local directory or in etc")
}

func main() {
	ctx := context.Background()

	if *rootNode == "" {
		exitf("must pass a non-blank --root")
	}

	conf, err := loadConfig()
	if err != nil {
		exitf(err.Error())
	}

	ex, err := explorer.New(*rootNode, conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	results, err := ex.Explore(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	l := explorer.List{}
	for _, node := range l.List(results.NetworkMap) {
		fmt.Println("Node: ", node.IP.String())
		fmt.Println("\tType: ", node.Type)
		if node.Error != nil {
			fmt.Println("\tError: ", node.Error)
		} else {
			fmt.Println("\tNeighbors:")
			for k, v := range node.Neighbors {
				fmt.Println("\t\t", k, ":", v.IP.String())
			}
		}
	}
}
