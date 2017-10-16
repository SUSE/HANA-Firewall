package model

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

const (
	FirewalldProtocolTCP = "tcp"
	FirewalldProtocolUDP = "udp"
)

// FirewalldService defines a service with its name, description, and ports.
type FirewalldService struct {
	ShortName   string          `xml:"short"`
	Description string          `xml:"description"`
	Ports       []FirewalldPort `xml:"port"`
}

// ToXML serialised service definition into a complete XML document that includes the XML header.
func (svc *FirewalldService) ToXML() string {
	tmp := struct {
		XMLName struct{} `xml:"service"` // the name of root element has to be "service"
		*FirewalldService
	}{FirewalldService: svc}

	out, err := xml.MarshalIndent(tmp, "", "    ")
	if err != nil {
		panic(err)
	}
	return xml.Header + string(out)
}

// String returns firewall service details in an easy to read, indented format.
func (svc *FirewalldService) String() string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%s - %s:\n", svc.ShortName, svc.Description))
	for _, port := range svc.Ports {
		out.WriteString(fmt.Sprintf("    Allow %s %d\n", port.Protocol, port.Port))
	}
	return out.String()
}

// FirewalldPort defines a port to be opened in a service.
type FirewalldPort struct {
	Port     int    `xml:"port,attr"`
	Protocol string `xml:"protocol,attr"`
}
