package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/ilmoraunio/pod-conftest-parser/babashka"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/russolsen/transit"
)

func debug(v interface{}) {
	fmt.Fprintf(os.Stderr, "debug: %+q\n", v)
}

func respond(message *babashka.Message, response interface{}) {
	buf := bytes.NewBufferString("")
	encoder := transit.NewEncoder(buf, false)

	if err := encoder.Encode(response); err != nil {
		babashka.WriteErrorResponse(message, err)
	} else {
		babashka.WriteInvokeResponse(message, string(buf.String()))
	}
}

func parseArgs(args string) ([]string, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()
	if err != nil {
		return []string{}, err
	}
	interfaceSlice := value.([]interface{})
	retval := make([]string, len(interfaceSlice))
	for i, v := range interfaceSlice {
		retval[i] = fmt.Sprint(v)
	}
	return retval, nil
}

func processMessage(message *babashka.Message) {
	switch message.Op {
	case "describe":
		babashka.WriteDescribeResponse(
			&babashka.DescribeResponse{
				Format: "transit+json",
				Namespaces: []babashka.Namespace{
					{
						Name: "pod.babashka.conftest-parser",
						Vars: []babashka.Var{
							{
								Name: "parse",
							},
						},
					},
				},
			})
	case "invoke":
		args, err := parseArgs(message.Args)
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			return
		}

		switch message.Var {
		case "pod.babashka.conftest-parser/parse":
			configs, err := parser.ParseConfigurations(args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			} else {
				respond(message, configs)
			}
		default:
			babashka.WriteErrorResponse(message, fmt.Errorf("Unknown var %s", message.Var))
		}
	default:
		babashka.WriteErrorResponse(message, fmt.Errorf("Unknown op %s", message.Op))
	}
}

func main() {
	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}
		processMessage(message)
	}
}
