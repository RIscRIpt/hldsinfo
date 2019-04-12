package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/RIscRIpt/hldsinfo/hldsinfo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <ip:port> [timeout (ms)]\n", os.Args[0])
		os.Exit(1)
	}

	serverAddress := os.Args[1]
	deadline := time.Now().Add(time.Duration(4) * time.Second)
	if len(os.Args) >= 3 {
		timeout, err := strconv.ParseInt(os.Args[2], 0, 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		if timeout <= 0 {
			deadline = time.Time{}
		} else {
			deadline = time.Now().Add(time.Duration(timeout) * time.Millisecond)
		}
	}

	info, err := hldsinfo.Get(serverAddress, deadline)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	infoJSON, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	fmt.Println(string(infoJSON))
}
