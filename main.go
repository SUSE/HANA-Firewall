package main

import (
	"bufio"
	"fmt"
	"github.com/HouzuoGuo/HANA-Firewall/generator"
	"github.com/HouzuoGuo/HANA-Firewall/model"
	"github.com/HouzuoGuo/HANA-Firewall/txtparser"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
		GenerateFirewalldServices()
	case "dry-run":
		DryRun()
	case "define-new-hana-service":
		CreateNewService()
	}
}

// readConfig reads HANA firewall configuration from /etc and return. If an error occurs, the program will exit.
func readConfig() (globalParams model.HANAGlobalParameters, services []model.HANAServiceDefinition) {
	// Read HANA firewall global config
	globalConf, err := txtparser.ParseSysconfigFile("/etc/sysconfig/hana-firewall", true)
	if err != nil {
		errorExit("Failed to create/open /etc/sysconfig/hana-firewall - %v", err)
		return
	}
	globalParams = model.HANAGlobalParameters{}
	globalParams.ReadFrom(globalConf)
	// Read HANA service definitions - all of them
	services = make([]model.HANAServiceDefinition, 0, 10)
	walkRoot := "/etc/hana-firewall"
	err = filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
		if path == walkRoot {
			// Move on from the directory
			return nil
		}
		service := model.HANAServiceDefinition{}
		serviceConf, err := txtparser.ParseSysconfigFile(path, true)
		if err != nil {
			log.Printf("GenerateFirewalldServices: skip definition file \"%s\" due to error - %v", path, err)
			// Move on
			return nil
		}
		service.ReadFrom(serviceConf)
		// Skip services with empty definition
		if len(service.TCP) > 0 || len(service.UDP) > 0 {
			service.FileBaseName = filepath.Base(path)
			services = append(services, service)
		}
		return nil
	})
	if err != nil {
		errorExit("Failed to read /etc/hana-firewall directory - %v", err)
		return
	}
	return
}

// GenerateFirewalldServices generates latest HANA service definition XML files for firewalld.
func GenerateFirewalldServices() {
	globalParams, services := readConfig()
	// Generate firewalld service definitions
	fw := generator.Firewalld{
		HANAGlobal:   globalParams,
		HANAServices: services,
	}
	firewalldServices, err := fw.GenerateConfig()
	if err != nil {
		errorExit("Failed to generate firewall config - %v", err)
		return
	}
	if len(firewalldServices) == 0 {
		errorExit("HANA instance number or service definitions are missing. Please check /etc/hana-firewall directory and /etc/sysconfig/hana-firewall file.")
		return
	}
	fmt.Printf("Generating %d services in /etc/firewalld/services:\n", len(firewalldServices))
	for _, svc := range firewalldServices {
		fmt.Println(svc.String())
		fmt.Println("----------------------------------------------------------")
	}
	// Write firewalld service definition XML
	if err := fw.WriteConfig("/etc/firewalld/services", firewalldServices); err != nil {
		errorExit("Failed to write XML files into /etc/firewalld/services - %v", err)
		return
	}
	fmt.Println(`All done!
Please restart firewalld service (systemctl restart firewalld.service) to make new HANA services visible.
Remember: transient firewall configuration are lost when restarting firewalld.service.`)
}

func DryRun() {
	globalParams, services := readConfig()
	// Generate firewalld service definitions
	fw := generator.Firewalld{
		HANAGlobal:   globalParams,
		HANAServices: services,
	}
	firewalldServices, err := fw.GenerateConfig()
	if err != nil {
		errorExit("Failed to generate firewall config - %v", err)
		return
	}
	for _, svc := range firewalldServices {
		fmt.Println(svc.String())
		fmt.Println("----------------------------------------------------------")
	}
	fmt.Println(`If you run "hana-firewall generate-firewalld-services", the services above will be made available in firewalld.`)
}

func CreateNewService() {
	stdin := bufio.NewReader(os.Stdin)
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("How would you like to name the new service? (e.g. \"database application support\"")
	name, _ := stdin.ReadString('\n')
	name = strings.TrimSpace(name)
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, '.') {
		errorExit("Sorry, the name may not contain slash or full-stop character.")
		return
	}
	if name == "" {
		errorExit("Sorry, you have to give the new service a name.")
		return
	}
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("Which TCP ports are used by the service? Use space to separate multiple ports. If there are none, simply press enter.")
	fmt.Println("For a special case, placeholder \"__INST_NUM__\" will be substituted by HANA instance numbers. and \"__INST_NUM+1__\" will be substituted by HANA instance number plus one.")
	fmt.Println("Examples: 3__INST_NUM__01 4__INST_NUM+1__02")
	tcpPortsStr, _ := stdin.ReadString('\n')
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("Which UDP ports are used by the service? Use space to separate multiple ports. If there are none, simply press enter.")
	fmt.Println("The special placeholders may also be used in these UDP ports.")
	udpPortsStr, _ := stdin.ReadString('\n')

	consecutiveSpaces := regexp.MustCompile("[[:space:]]+")
	tcpPorts := consecutiveSpaces.Split(strings.TrimSpace(tcpPortsStr), -1)
	udpPorts := consecutiveSpaces.Split(strings.TrimSpace(udpPortsStr), -1)
	if len(tcpPorts) == 0 && len(udpPorts) == 0 {
		errorExit("Sorry, the service must have at least one TCP or UDP port defined.")
		return
	}

	service := model.HANAServiceDefinition{
		FileBaseName: name,
		TCP:          tcpPorts,
		UDP:          udpPorts,
	}
	filePath := path.Join("/etc/hana-firewall/", name)
	serviceConf, err := txtparser.ParseSysconfigFile(filePath, true)
	if err != nil {
		errorExit("Failed to create service definition file at \"%s\": %v", filePath, err)
		return
	}
	service.WriteInto(serviceConf)
	if err := ioutil.WriteFile(filePath, []byte(serviceConf.ToText()), 0600); err != nil {
		errorExit("Failed to create service definition file at \"%s\": %v", filePath, err)
		return
	}
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("All done! Remember to run \"hana-firewall generate-firewalld-services\" to make use of the new service.")
}
