package main

import (
	"fmt"
	"os"
	"strings"
)

// printHelpAndExit prints usage help and then exits the program.
func printHelpAndExit(exitStatus int) {
	fmt.Println(`hana-firewall: helps to generate HANA network service definitions for firewalld.
Usage:
	# hana-firewall generate-firewalld-services
		Generate firewalld service XML files according to HANA service definitions.
		Previously generated XML files will be overwritten.
	# hana-firewall dry-run
		Display the service name and port numbers that will be generated in firewalld service XML files.
	# hana-firewall define-new-hana-service
		Interactively create a new HANA network service definition.
	# hana-firewall help
		Display this help message.
`)
	os.Exit(exitStatus)
}

// cliArg returns the i-th command line parameter, or an empty string if the parameter is not specified.
func cliArg(i int) string {
	if len(os.Args) >= i+1 {
		return os.Args[i]
	}
	return ""
}

// errorExit prints out a message to standard error and then exits the program with status 1.
func errorExit(template string, stuff ...interface{}) {
	fmt.Fprintf(os.Stderr, template+"\n", stuff...)
	os.Exit(1)
}
func main() {
	if arg1 := cliArg(1); arg1 == "" || strings.Contains(arg1, "help") {
		printHelpAndExit(0)
	}
	// All other actions require root privilege
	if os.Geteuid() != 0 {
		errorExit("Please run hana-firewall with root privilege.")
		return
	}
	switch cliArg(1) {
	case "generate-firewalld-services":
	case "dry-run":
	case "create-new-services":
	}
}
