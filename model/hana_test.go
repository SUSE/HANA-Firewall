package model

import (
	"github.com/SUSE/HANA-Firewall/txtparser"
	"reflect"
	"testing"
)

var globalParams = HANAGlobalParameters{
	InstanceNumbers: []string{"00", "01"},
}

var definition = HANAServiceDefinition{
	FileBaseName: "Database Client",
	TCP:          []string{"1__INST_NUM__2", "34", "5__INST_NUM+1__6"},
	UDP:          []string{"1__INST_NUM__2", "34", "5__INST_NUM+1__6"},
}

func TestMakeFirewalldService(t *testing.T) {
	shortName, svc, err := globalParams.MakeFirewalldService(&definition)
	if err != nil {
		t.Fatal(err)
	}
	if shortName != "database-client" {
		t.Fatal(shortName)
	}

	match := FirewalldService{
		ShortName:   "database-client",
		Description: "Database Client",
		Ports: []FirewalldPort{
			{Protocol: "tcp", Port: 34},
			{Protocol: "tcp", Port: 1002},
			{Protocol: "tcp", Port: 1012},
			{Protocol: "tcp", Port: 5016},
			{Protocol: "tcp", Port: 5026},

			{Protocol: "udp", Port: 34},
			{Protocol: "udp", Port: 1002},
			{Protocol: "udp", Port: 1012},
			{Protocol: "udp", Port: 5016},
			{Protocol: "udp", Port: 5026},
		},
	}
	// Deserialise XML into structure and verify
	if !reflect.DeepEqual(svc, match) {
		t.Fatalf("%+v", svc)
	}

}

func TestHANAGlobalParametersSysconfig(t *testing.T) {
	sample := `## Path:        Network/Firewall/HANA Firewall/Global Configuration
## Type:        string
## Default:     ""
#
# Space-separated list of HANA system instance numbers that will participate
# firewall setup. For example, if there are three HANA installations called
# "PRD01, "TST02", and "DEV03", then write down "01 02 03" in the value.
#
# The instance numbers will take part in generating many firewall service
# definitions.
#
HANA_INSTANCE_NUMBERS="00 01"
`
	conf, err := txtparser.ParseSysconfig(sample)
	if err != nil {
		t.Fatal(err)
	}
	var global HANAGlobalParameters
	global.ReadFrom(conf)
	if !reflect.DeepEqual(global.InstanceNumbers, []string{"00", "01"}) {
		t.Fatalf("%+v", global)
	}

	global.InstanceNumbers = []string{"02", "03"}
	global.WriteInto(conf)
	match := `## Path:        Network/Firewall/HANA Firewall/Global Configuration
## Type:        string
## Default:     ""
#
# Space-separated list of HANA system instance numbers that will participate
# firewall setup. For example, if there are three HANA installations called
# "PRD01, "TST02", and "DEV03", then write down "01 02 03" in the value.
#
# The instance numbers will take part in generating many firewall service
# definitions.
#
HANA_INSTANCE_NUMBERS="02 03"
`
	if s := conf.ToText(); s != match {
		t.Fatalf("\n%v\n%v\n", []byte(s), []byte(match))
	}
}

func TestHANAServiceDefinition(t *testing.T) {
	sample := `# HANA special support
# The ports should be used in rare technical support scenarios. See HANA administration guide for more details.

TCP="3__INST_NUM__09 1000"
UDP="3__INST_NUM__09 2000"
`
	conf, err := txtparser.ParseSysconfig(sample)
	if err != nil {
		t.Fatal(err)
	}
	var def HANAServiceDefinition
	def.ReadFrom(conf)
	if !reflect.DeepEqual(def.TCP, []string{"3__INST_NUM__09", "1000"}) {
		t.Fatalf("%+v", def)
	}
	if !reflect.DeepEqual(def.UDP, []string{"3__INST_NUM__09", "2000"}) {
		t.Fatalf("%+v", def)
	}

	def.TCP = []string{"3000", "3001"}
	def.UDP = []string{"4000", "4001"}
	def.WriteInto(conf)
	match := `# HANA special support
# The ports should be used in rare technical support scenarios. See HANA administration guide for more details.

TCP="3000 3001"
UDP="4000 4001"
`
	if s := conf.ToText(); s != match {
		t.Fatalf("\n%v\n%v\n", []byte(s), []byte(match))
	}
}

func TestHANAServiceDefinition_GetShortName(t *testing.T) {
	def := HANAServiceDefinition{FileBaseName: "/a?V&XDFn9_QW_.{:}|"}
	name := def.GetShortName()
	if name != "-a-v-xdfn9-qw------" {
		t.Fatal(name)
	}
}

func TestUniqueSortedInts(t *testing.T) {
	in := []int{0, 1, 5, 2, 5, 2, 6}
	out := UniqueSortedInts(in)
	if !reflect.DeepEqual(out, []int{0, 1, 2, 5, 6}) {
		t.Fatal(out)
	}
}
