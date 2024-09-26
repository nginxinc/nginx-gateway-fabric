package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	crossplane "github.com/nginxinc/nginx-go-crossplane"
)

// This binary accepts a single argument, the path of the base nginx config, and prints out the JSON representation
// of the full nginx config, in crossplane format.
// See https://github.com/nginxinc/nginx-go-crossplane for more info.
func main() {
	if len(os.Args) != 2 {
		panic(errors.New("must have exactly one argument, the path of the base nginx config"))
	}

	path := os.Args[1]

	payload, err := crossplane.Parse(path, &crossplane.ParseOptions{})
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
