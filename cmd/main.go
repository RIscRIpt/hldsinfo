package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/RIscRIpt/hldsinfo/hldsinfo"
)

var timeout = flag.Int64("t", 4000, "timeout")

func main() {
	flag.Parse()

	if flag.NArg() <= 0 {
		fmt.Printf("usage: %s <ip:port> ... [-t <timeout (ms)>]\n", os.Args[0])
		os.Exit(1)
	}

	fetcher := hldsinfo.NewFetcher(time.Duration(*timeout) * time.Millisecond)
	for _, serverAddress := range flag.Args() {
		fetcher.Fetch(serverAddress)
	}

	infos := fetcher.Get()
	fmt.Println("[")
	i := 0
	for _, info := range infos {
		var s string
		if info != nil {
			infoJSON, err := json.MarshalIndent(info, "    ", "    ")
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			s = string(infoJSON)
		} else {
			s = "{}"
		}
		if i < len(infos)-1 {
			fmt.Printf("    %s,\n", s)
		} else {
			fmt.Printf("    %s\n", s)
		}
		i++
	}
	fmt.Println("]")
}
