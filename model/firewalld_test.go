package model

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestFirewalldXML(t *testing.T) {
	sample := `<?xml version="1.0" encoding="utf-8"?>
<service>
    <short>This is short name</short>
    <description>This is description</description>
    <port protocol="tcp" port="80"/>
    <port protocol="tcp" port="443"/>
    <port protocol="tcp" port="88"/>
    <port protocol="udp" port="88"/>
</service>`
	match := FirewalldService{
		ShortName:   "This is short name",
		Description: "This is description",
		Ports: []FirewalldPort{
			{Protocol: "tcp", Port: 80},
			{Protocol: "tcp", Port: 443},
			{Protocol: "tcp", Port: 88},
			{Protocol: "udp", Port: 88},
		},
	}

	// Deserialise XML into structure
	var elem FirewalldService
	if err := xml.Unmarshal([]byte(sample), &elem); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(elem, match) {
		t.Fatalf("%+v", elem)
	}

	// Serialise structure into XML and match again
	toXML := match.ToXML()
	elem = FirewalldService{}
	if err := xml.Unmarshal([]byte(toXML), &elem); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(elem, match) {
		t.Fatalf("\n%+v\n%+v\n", elem, match)
	}

	// Format for readability
	matchStr := `This is short name - This is description:
    Allow tcp 80
    Allow tcp 443
    Allow tcp 88
    Allow udp 88
`
	if s := match.String(); s != matchStr {
		t.Fatalf("\n%s\n%s\n%v\n%v\n", s, matchStr, []byte(s), []byte(matchStr))
	}
}
