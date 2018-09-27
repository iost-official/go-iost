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
	cid := d.h.db.MGet(DHCPTable, url)
	if cid == "n" {
		return ""
	}
	return cid
}

// URLOwner find owner of url
func (d *DHCP) URLOwner(url string) string {
	owner := d.h.db.MGet(DHCPOwnerTable, url)
	if owner == "n" {
		return ""
	}
	return owner
}

// URLTransfer give url to another id
func (d *DHCP) URLTransfer(url, to string) {
	d.h.db.MPut(DHCPOwnerTable, url, to)
}

// URL get url of cid
func (d *DHCP) URL(cid string) string {
	domain := d.h.db.MGet(DHCPRTable, cid)
	if domain == "n" {
		return ""
	}
	return domain
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
	d.h.db.MPut(DHCPTable, url, cid)
	d.h.db.MPut(DHCPRTable, cid, url)
	d.h.db.MPut(DHCPOwnerTable, url, owner)
}

// RemoveLink remove a url
func (d *DHCP) RemoveLink(url, cid string) {
	d.h.db.MDel(DHCPRTable, cid)
	d.h.db.MDel(DHCPTable, url)
	d.h.db.MDel(DHCPOwnerTable, url)
}
