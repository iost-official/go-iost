package host

import (
	"strings"
)

// const table name
const (
	DNSTable      = "dns_table"
	DNSRTable     = "dns_revert_table"
	DNSOwnerTable = "dns_owner_table"
)

// DNS dns server handler
type DNS struct {
	h *Host
}

// NewDNS make a dns
func NewDNS(h *Host) DNS {
	return DNS{
		h: h,
	}
}

// ContractID find cid from url
func (d *DNS) ContractID(url string) string {
	cid, _ := d.h.GlobalMapGet("domain.iost", DNSTable, url)
	if s, ok := cid.(string); ok {
		return s
	}
	return ""
}

// URLOwner find owner of url
func (d *DNS) URLOwner(url string) string {
	owner, _ := d.h.GlobalMapGet("domain.iost", DNSOwnerTable, url)
	if s, ok := owner.(string); ok {
		return s
	}
	return ""
}

// URLTransfer give url to another id
func (d *DNS) URLTransfer(url, to string) {
	d.h.MapPut(DNSOwnerTable, url, to)
}

// URL git url of cid
func (d *DNS) URL(cid string) string {
	domain, _ := d.h.GlobalMapGet("domain.iost", DNSRTable, cid)
	if s, ok := domain.(string); ok {
		return s
	}
	return ""
}

// IsDomain determine if s is a domain
func (d *DNS) IsDomain(s string) bool {
	if strings.HasPrefix(s, "Contract") || strings.HasSuffix(s, ".iost") {
		return false
	}
	return true
}

// WriteLink add url and url owner to contract
func (d *DNS) WriteLink(url, cid, owner string) {
	d.h.MapPut(DNSTable, url, cid)
	d.h.MapPut(DNSRTable, cid, url)
	d.h.MapPut(DNSOwnerTable, url, owner)
}

// RemoveLink remove a url
func (d *DNS) RemoveLink(url, cid string) {
	d.h.MapDel(DNSRTable, cid)
	d.h.MapDel(DNSTable, url)
	d.h.MapDel(DNSOwnerTable, url)
}
