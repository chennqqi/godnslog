package fingerprint

import "net"

var cloudCIDRs = []struct {
	name string
	cidr *net.IPNet
}{
	// Cloudflare
	{"Cloudflare", mustCIDR("103.21.244.0/22")},
	{"Cloudflare", mustCIDR("103.22.200.0/22")},
	{"Cloudflare", mustCIDR("104.16.0.0/13")},
	{"Cloudflare", mustCIDR("131.0.72.0/22")},
	{"Cloudflare", mustCIDR("172.64.0.0/13")},
	// AWS
	{"AWS", mustCIDR("13.32.0.0/15")},
	{"AWS", mustCIDR("52.0.0.0/15")},
	{"AWS", mustCIDR("54.0.0.0/14")},
	{"AWS", mustCIDR("52.94.0.0/15")},
	{"AWS", mustCIDR("52.119.192.0/20")},
	// Google Cloud
	{"Google Cloud", mustCIDR("34.0.0.0/15")},
	{"Google Cloud", mustCIDR("35.184.0.0/14")},
	{"Google Cloud", mustCIDR("35.192.0.0/14")},
	{"Google Cloud", mustCIDR("35.224.0.0/13")},
	// Azure
	{"Azure", mustCIDR("13.64.0.0/11")},
	{"Azure", mustCIDR("13.96.0.0/13")},
	{"Azure", mustCIDR("20.0.0.0/10")},
	{"Azure", mustCIDR("40.74.0.0/15")},
	{"Azure", mustCIDR("52.128.0.0/12")},
	// Alibaba Cloud
	{"Alibaba", mustCIDR("8.128.0.0/10")},
	{"Alibaba", mustCIDR("47.88.0.0/14")},
	{"Alibaba", mustCIDR("47.96.0.0/12")},
	// Oracle Cloud (OCI)
	{"Oracle", mustCIDR("129.146.0.0/16")},
	{"Oracle", mustCIDR("130.61.0.0/16")},
	// DigitalOcean
	{"DigitalOcean", mustCIDR("159.65.0.0/16")},
	{"DigitalOcean", mustCIDR("165.227.0.0/16")},
	{"DigitalOcean", mustCIDR("138.68.0.0/16")},
	// Vultr
	{"Vultr", mustCIDR("45.32.0.0/16")},
	{"Vultr", mustCIDR("108.61.0.0/16")},
	// Linode
	{"Linode", mustCIDR("45.33.0.0/16")},
	{"Linode", mustCIDR("139.162.0.0/16")},
	// Hetzner
	{"Hetzner", mustCIDR("5.9.0.0/16")},
	{"Hetzner", mustCIDR("78.46.0.0/15")},
	{"Hetzner", mustCIDR("88.198.0.0/16")},
	// OVH
	{"OVH", mustCIDR("51.38.0.0/16")},
	{"OVH", mustCIDR("51.75.0.0/16")},
	{"OVH", mustCIDR("141.94.0.0/16")},
	// Tencent Cloud
	{"Tencent", mustCIDR("81.68.0.0/14")},
	{"Tencent", mustCIDR("101.32.0.0/14")},
	{"Tencent", mustCIDR("134.175.0.0/16")},
	// Huawei Cloud
	{"Huawei", mustCIDR("49.4.0.0/14")},
	{"Huawei", mustCIDR("121.36.0.0/15")},
	{"Huawei", mustCIDR("122.9.0.0/16")},
}
