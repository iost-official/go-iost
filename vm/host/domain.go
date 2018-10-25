package host

import (
	"strings"
)

// const table name
const (
	DHCPTable      = "dhcp_table"
	DHCPRTable     = "dhcp_revert_table"
	DHCPOwnerTable = "dhcp_owner_table"
)

// DHCP dhcp server handler
type DHCP struct {
	h *Host
}

// NewDHCP make a dhcp
func NewDHCP(h *Host) DHCP {
	return DHCP{
		h: h,
	}
}

// ContractID find cid from url
func (d *DHCP) ContractID(url string) string {
	cid, _ := d.h.GlobalMapGet("iost.domain", DHCPTable, url)
	if s, ok := cid.(string); ok {
		return s
	}
	return ""
}

// URLOwner find owner of url
func (d *DHCP) URLOwner(url string) string {
	owner, _ := d.h.GlobalMapGet("iost.domain", DHCPOwnerTable, url)
	if s, ok := owner.(string); ok {
		return s
	}
	return ""
}

// URLTransfer give url to another id
func (d *DHCP) URLTransfer(url, to string) {
	d.h.MapPut(DHCPOwnerTable, url, to)
}

// URL git url of cid
func (d *DHCP) URL(cid string) string {
	domain, _ := d.h.GlobalMapGet("iost.domain", DHCPRTable, cid)
	if s, ok := domain.(string); ok {
		return s
	}
	return ""
}

// IsDomain determine if s is a domain
func (d *DHCP) IsDomain(s string) bool {
	if strings.HasPrefix(s, "Contract") || strings.HasPrefix(s, "iost.") {
		return false
	}
	return true
}

// WriteLink add url and url owner to contract
func (d *DHCP) WriteLink(url, cid, owner string) {
	d.h.MapPut(DHCPTable, url, cid)
	d.h.MapPut(DHCPRTable, cid, url)
	d.h.MapPut(DHCPOwnerTable, url, owner)
}

// RemoveLink remove a url
func (d *DHCP) RemoveLink(url, cid string) {
	d.h.MapDel(DHCPRTable, cid)
	d.h.MapDel(DHCPTable, url)
	d.h.MapDel(DHCPOwnerTable, url)
}
