package main

import (
	"encoding/json"
	"fmt"

	"github.com/kiasaki/go-sdc"
)

func main() {
	client := sdc.DefaultClient()

	machines, err := client.ListMachines()
	outputErrorOrResult(err, machines)
}

func outputErrorOrResult(err error, data interface{}) {
	if err != nil {
		if _, ok := err.(sdc.SDCError); ok {
			output(err)
		} else {
			output(sdc.NewSDCError("Unknown", err.Error()))
		}
	} else {
		output(data)
	}
}

func output(data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%q\n", err.Error())
	} else {
		fmt.Println(string(bytes))
	}
}
