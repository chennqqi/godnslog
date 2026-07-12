package fingerprint

import "net"

var scannerCIDRs = []struct {
	name string
	cidr *net.IPNet
}{
	// ProjectDiscovery / Nuclei cloud infrastructure
	{"ProjectDiscovery", mustCIDR("45.33.0.0/16")},
	// Shodan
	{"Shodan", mustCIDR("141.98.10.0/24")},
	{"Shodan", mustCIDR("141.98.11.0/24")},
	// Censys
	{"Censys", mustCIDR("162.142.125.0/24")},
	{"Censys", mustCIDR("167.94.138.0/24")},
	{"Censys", mustCIDR("167.94.145.0/24")},
	{"Censys", mustCIDR("167.94.146.0/24")},
	// BinaryEdge
	{"BinaryEdge", mustCIDR("185.180.13.0/24")},
	// Internet Archive
	{"InternetArchive", mustCIDR("207.241.224.0/20")},
	// ONYPHE
	{"ONYPHE", mustCIDR("193.218.118.0/24")},
	// Netlas
	{"Netlas", mustCIDR("185.38.129.0/24")},
	// ZoomEye
	{"ZoomEye", mustCIDR("185.236.112.0/24")},
	// Fofa / 华顺
	{"Fofa", mustCIDR("106.75.0.0/16")},
	// 360 Quake
	{"Quake", mustCIDR("180.163.50.0/24")},
	{"Quake", mustCIDR("101.126.0.0/16")},
	// 知道创宇
	{"KnownSec", mustCIDR("123.56.0.0/16")},
	{"KnownSec", mustCIDR("47.92.0.0/16")},
	// TODO: add more scanner CIDRs
}

var scannerUAKeywords = []struct {
	keyword string
	name    string
}{
	{"nuclei", "Nuclei"},
	{"nessus", "Nessus"},
	{"acunetix", "Acunetix"},
	{"awvs", "AWVS"},
	{"netsparker", "Netsparker"},
	{"python-requests", "Python Requests"},
	{"zgrab", "ZGrab"},
	{"masscan", "Masscan"},
	{"nmap", "Nmap"},
	{"whatweb", "WhatWeb"},
	{"wpscan", "WPScan"},
	{"sqlmap", "SQLMap"},
	{"gobuster", "GoBuster"},
	{"dirsearch", "Dirsearch"},
	{"hydra", "Hydra"},
	{"burpsuite", "BurpSuite"},
	{"postman", "Postman"},
}

func mustCIDR(s string) *net.IPNet {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		panic("invalid CIDR: " + s + ": " + err.Error())
	}
	return cidr
}
