package net

type Network struct {
	Name        string
	ID          uint64
	SelfLink    string
	IPv4Range   string // for legacy/auto mode
	Mode        string // "AUTO", "CUSTOM", "LEGACY"
	GatewayIPv4 string
}

type Subnet struct {
	Name        string
	Region      string
	IPCidrRange string
	Gateway     string
	Network     string // link to network
}

type Firewall struct {
	Name      string
	Network   string
	Direction string // INGRESS, EGRESS
	Priority  int64
	Action    string // ALLOW, DENY
	Source    string // Ranges or Tags
	Target    string // Ranges or Tags
}
