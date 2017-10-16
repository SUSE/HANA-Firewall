package model

import (
	"bytes"
	"fmt"
	"github.com/HouzuoGuo/HANA-Firewall/txtparser"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	/*
		InstanceNumberSubstitutionMagic is a substring that appears among HANA networks service port definitions.
		The substring serves as a placeholder, the actual port number will be calculated by substituting the placeholder
		by a HANA instance number. For example, given a port definition of "1__INST_NUM__2" and an instance number of 00,
		the calculated port number will be "1002".
	*/
	InstanceNumberSubstitutionMagic = "__INST_NUM__"

	/*
		InstanceNumberPlusOneSubstitutionMagic is a substring that appears among HANA networks service port definitions.
		Similar to InstanceNumberSubstitutionMagic, this placeholder will be substituted by instance number plus one.
	*/
	InstanceNumberPlusOneSubstitutionMagic = "__INST_NUM+1__"

	HANAServiceDefinitionTCPKey  = "TCP"
	HANAServiceDefinitionUDPKey  = "UDP"
	HANAGlobalInstanceNumbersKey = "HANA_INSTANCE_NUMBERS"
)

// HANAServiceDefinition is a HANA network service definition written in a sysconfig-style text file.
type HANAServiceDefinition struct {
	FileBaseName string   // FileBaseName is the base name of service definition file.
	TCP          []string // TCP port numbers, each one may include the instance number substitution magic.
	UDP          []string // UDP port numbers, each one may include the instance number substitution magic.
}

// GetShortName returns a linted "short name" that identifies a Firewalld service and its XML file.
func (def *HANAServiceDefinition) GetShortName() string {
	var ret bytes.Buffer
	// Only retain numbers and letters, turn letters lower case.
	for _, c := range def.FileBaseName {
		if unicode.IsNumber(c) {
			ret.WriteRune(c)
		} else if unicode.IsLetter(c) {
			ret.WriteRune(unicode.ToLower(c))
		} else {
			ret.WriteRune('-')
		}
	}
	return ret.String()
}

// ReadFromText reads definition content from a sysconfig-style text file.
func (def *HANAServiceDefinition) ReadFrom(txt *txtparser.Sysconfig) {
	def.TCP = txt.GetStringArray(HANAServiceDefinitionTCPKey, []string{})
	def.UDP = txt.GetStringArray(HANAServiceDefinitionUDPKey, []string{})
}

// WriteInto overwrites keys and values of text file with the current definition content.
func (def *HANAServiceDefinition) WriteInto(txt *txtparser.Sysconfig) {
	txt.SetStringArray(HANAServiceDefinitionTCPKey, def.TCP)
	txt.SetStringArray(HANAServiceDefinitionUDPKey, def.UDP)
}

// HANAGlobalParameters are settings that come from /etc/sysconfig/hana-firewall.
type HANAGlobalParameters struct {
	InstanceNumbers []string
}

func (global *HANAGlobalParameters) ReadFrom(txt *txtparser.Sysconfig) {
	global.InstanceNumbers = txt.GetStringArray(HANAGlobalInstanceNumbersKey, []string{})
}

func (global *HANAGlobalParameters) WriteInto(txt *txtparser.Sysconfig) {
	txt.SetStringArray(HANAGlobalInstanceNumbersKey, global.InstanceNumbers)
}

/*
GetPortNumbers returns actual service port numbers calculated by expanding definition string with instance number
parameter. An error will be returned only if there is a number formatting.
*/
func (global *HANAGlobalParameters) GetPortNumbers(portDefinition string) (ret []int, err error) {
	ret = make([]int, 0, 10)
	expandedPorts := make([]string, 0, 10)
	for _, instNumStr := range global.InstanceNumbers {
		instancePort := portDefinition
		// Replace magic strings among the definition by instance number string
		if strings.Contains(instancePort, InstanceNumberSubstitutionMagic) {
			instancePort = strings.Replace(instancePort, InstanceNumberSubstitutionMagic, instNumStr, -1)
		}
		if strings.Contains(instancePort, InstanceNumberPlusOneSubstitutionMagic) {
			// Convert instance number string into integer, plus one, and add padding zero on the left.
			instNum, err := strconv.Atoi(instNumStr)
			if err != nil {
				return ret, fmt.Errorf("HANAGlobalParameters.GetPortNumbers: from global parameters, an instance number \"%s\" is not a valid integer", instNumStr)
			}
			instancePort = strings.Replace(instancePort, InstanceNumberPlusOneSubstitutionMagic, fmt.Sprintf("%.2d", instNum+1), -1)
		}
		expandedPorts = append(expandedPorts, instancePort)
	}
	// Turn expanded port strings into integers
	for _, portStr := range expandedPorts {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return ret, fmt.Errorf("HANAGlobalParameters.GetPortNumbers: failed to interpret port number \"%s\" while expanding \"%s\"", portStr, portDefinition)
		}
		ret = append(ret, port)
	}
	return
}

// MakeFirewalldService generates firewalld service definition for a single HANA service definition.
func (global *HANAGlobalParameters) MakeFirewalldService(def *HANAServiceDefinition) (serviceShortName string, svc FirewalldService, err error) {
	serviceShortName = def.GetShortName()
	// Calculate actual TCP and UDP port numbers
	tcpPorts := make([]int, 0, 10)
	udpPorts := make([]int, 0, 10)
	for _, portDefinition := range def.TCP {
		var actualPortNumbers []int
		actualPortNumbers, err = global.GetPortNumbers(portDefinition)
		if err != nil {
			return
		}
		tcpPorts = append(tcpPorts, actualPortNumbers...)
	}
	for _, portDefinition := range def.UDP {
		var actualPortNumbers []int
		actualPortNumbers, err = global.GetPortNumbers(portDefinition)
		if err != nil {
			return
		}
		udpPorts = append(udpPorts, actualPortNumbers...)
	}
	ports := make([]FirewalldPort, 0, 10)
	for _, port := range UniqueSortedInts(tcpPorts) {
		ports = append(ports, FirewalldPort{
			Port:     port,
			Protocol: FirewalldProtocolTCP,
		})
	}
	for _, port := range UniqueSortedInts(udpPorts) {
		ports = append(ports, FirewalldPort{
			Port:     port,
			Protocol: FirewalldProtocolUDP,
		})
	}

	svc = FirewalldService{
		ShortName:   serviceShortName,
		Description: def.FileBaseName,
		Ports:       ports,
	}
	return
}

// UniqueSortedInts returns unique integers among the input, sorted in ascending order.
func UniqueSortedInts(in []int) (out []int) {
	uniq := map[int]struct{}{}
	for _, i := range in {
		uniq[i] = struct{}{}
	}
	out = make([]int, 0, len(uniq))
	for i, _ := range uniq {
		out = append(out, i)
	}
	sort.Ints(out)
	return
}
