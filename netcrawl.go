package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/johnsiilver/netcrawl/explorer"

	"github.com/mewbak/gopass"
)

var (
	rootNode = flag.String("root", "", "The IP/Hostname of the root device")
)

func exitf(s string, a ...interface{}) {
	fmt.Printf(s, a...)
	fmt.Println()
	os.Exit(1)
}

func main() {
	ctx := context.Background()

	if *rootNode == "" {
		exitf("must pass a non-blank --root")
	}

	user, err := gopass.GetPass("User: ")
	if err != nil {
		exitf(err.Error())
	}
	pass, err := gopass.GetPass("Pass: ")
	if err != nil {
		exitf(err.Error())
	}

	ex, err := explorer.New(*rootNode, user, pass)
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
		}else{
			fmt.Println("\tNeighbors:")
			for k, v := range node.Neighbors {
				fmt.Println("\t\t", k, ":", v.IP.String())
			}
		}
	}
}
