package model

const (
	InstanceNumberSubstitutionMagic = "__INST_NUM__"
)

// ServiceDefinition is a HANA network service definition, consisting of TCP and UDP port definitions.
type ServiceDefinition struct {
	TCP []string
	UDP []string
}

type Parameters struct {
	InstanceNumbers []string
}
