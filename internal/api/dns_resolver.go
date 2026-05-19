package api

import "net"

// NetDNSResolver implements DNSResolver using the Go net package.
type NetDNSResolver struct{}

// LookupMX resolves MX records for the given host.
func (r *NetDNSResolver) LookupMX(host string) ([]string, error) {
	mxRecords, err := net.LookupMX(host)
	if err != nil {
		return nil, err
	}
	var hosts []string
	for _, mx := range mxRecords {
		hosts = append(hosts, mx.Host)
	}
	return hosts, nil
}

// LookupTXT resolves TXT records for the given host.
func (r *NetDNSResolver) LookupTXT(host string) ([]string, error) {
	return net.LookupTXT(host)
}
