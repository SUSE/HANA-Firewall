package generator

import (
	"encoding/xml"
	"github.com/HouzuoGuo/HANA-Firewall/model"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestFirewalld(t *testing.T) {
	fw := Firewalld{
		HANAGlobal: model.HANAGlobalParameters{
			InstanceNumbers: []string{"00", "01"},
		},
		HANAServices: []model.HANAServiceDefinition{
			{
				FileBaseName: "Database Client",
				TCP:          []string{"1__INST_NUM__00", "200"},
			},
			{
				FileBaseName: "B^$&VGDF#C$",
				UDP:          []string{"3__INST_NUM+1__00", "400"},
			},
		},
	}
	// Convert to firewalld services
	services, err := fw.GenerateConfig()
	if err != nil {
		t.Fatal(err)
	}
	match := map[string]model.FirewalldService{
		"database-client": {
			ShortName:   "database-client",
			Description: "Database Client",
			Ports: []model.FirewalldPort{
				{Port: 200, Protocol: "tcp"},
				{Port: 10000, Protocol: "tcp"},
				{Port: 10100, Protocol: "tcp"},
			},
		},
		"b---vgdf-c-": {
			ShortName:   "b---vgdf-c-",
			Description: "B^$&VGDF#C$",
			Ports: []model.FirewalldPort{
				{Port: 400, Protocol: "udp"},
				{Port: 30100, Protocol: "udp"},
				{Port: 30200, Protocol: "udp"},
			},
		},
	}
	if !reflect.DeepEqual(match, services) {
		t.Fatalf("\n%+v\n%+v\n", match, services)
	}

	// Write firewalld services into XML files
	dest, err := ioutil.TempDir("", "hana-firewall-TestFirewalld")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dest)
	if err := fw.WriteConfig(dest, services); err != nil {
		t.Fatal(err)
	}

	// Read back the XML files and verify
	databaseClientXML, err := ioutil.ReadFile(path.Join(dest, "database-client.xml"))
	if err != nil {
		t.Fatal(err)
	}
	theOtherXML, err := ioutil.ReadFile(path.Join(dest, "b---vgdf-c-.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var dbService, theOtherService model.FirewalldService
	if err := xml.Unmarshal(databaseClientXML, &dbService); err != nil {
		t.Fatal(err)
	}
	if err := xml.Unmarshal(theOtherXML, &theOtherService); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(dbService, match["database-client"]) {
		t.Fatalf("%+v", dbService)
	}
	if !reflect.DeepEqual(theOtherService, match["b---vgdf-c-"]) {
		t.Fatalf("%+v", dbService)
	}
}
