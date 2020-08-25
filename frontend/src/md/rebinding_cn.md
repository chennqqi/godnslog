CEYE DNS Rebinding Server

## 1. What is DNS rebinding
This is wikipedia explanation.

DNS rebinding is a form of computer attack. In this attack, a malicious web page causes visitors to run a client-side script that attacks machines elsewhere on the network. In theory, the same-origin policy prevents this from happening: client-side scripts are only allowed to access content on the same host that served the script. Comparing domain names is an essential part of enforcing this policy, so DNS rebinding circumvents this protection by abusing the Domain Name System (DNS).

This attack can be used to breach a private network by causing the victim's web browser to access machines at private IP addresses and return the results to the attacker. It can also be employed to use the victim machine for spamming, distributed denial-of-service attacks or other malicious activities.

## 2. How DNS rebinding works
The attacker registers a domain (such as attacker.com) and delegates it to a DNS server under the attacker's control. The server is configured to respond with a very short time to live (TTL) record, preventing the response from being cached. When the victim browses to the malicious domain, the attacker's DNS server first responds with the IP address of a server hosting the malicious client-side code. For instance, they could point the victim's browser to a website that contains malicious JavaScript or Flash scripts that are intended to execute on the victim's computer.

The malicious client-side code makes additional accesses to the original domain name (such as attacker.com). These are permitted by the same-origin policy. However, when the victim's browser runs the script it makes a new DNS request for the domain, and the attacker replies with a new IP address. For instance, they could reply with an internal IP address or the IP address of a target somewhere else on the Internet.

## 3. How to use CEYE DNS rebinding
Access the profile page, add the dns address you need to resolve, as shown below 

![](https://images.seebug.org/ceye/dnsrebinding.jpeg)

If your identifier is abcdef.ceye.io, then your DNS rebinding host is r.abcdef.ceye.io. You can use r.abcdef.ceye.io or *.r.abcdef.ceye.io to explore.

For example, use nslookup multiple times, the dns answer section will randomly return one of them

	$ nslookup r.abcdef.ceye.io
	Server:     8.8.8.8
	Address:    8.8.8.8#53
	Non-authoritative answer:
	Name:    r.abcdef.ceye.io
	Address: 127.0.0.1

	$ nslookup r.abcdef.ceye.io
	Server:     8.8.8.8
	Address:    8.8.8.8#53
	Non-authoritative answer:
	Name:    r.abcdef.ceye.io
	Address: 192.168.0.1

	$ nslookup r.abcdef.ceye.io
	Server:      8.8.8.8
	Address:    8.8.8.8#53
	Non-authoritative answer:
	Name:    r.abcdef.ceye.io
	Address: 127.0.0.1