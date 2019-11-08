package host

import (
	"strings"

	"github.com/iost-official/go-iost/ilog"
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
	_, err := d.h.MapPut(DNSOwnerTable, url, to)
	if err != nil {
		ilog.Errorf("url transfer mapPut failed. err = %v", err)
	}
}

// URL get url of cid
func (d *DNS) URL(cid string) string {
	domain, _ := d.h.GlobalMapGet("domain.iost", DNSRTable, cid)
	if s, ok := domain.(string); ok {
		return s
	}
	return ""
}

// IsDomain determine if s is a domain
func (d *DNS) IsDomain(s string) bool {
	return !strings.HasPrefix(s, "Contract")
}

// WriteLink add url and url owner to contract
func (d *DNS) WriteLink(url, cid, owner string) {
	_, err0 := d.h.MapPut(DNSTable, url, cid)
	_, err1 := d.h.MapPut(DNSRTable, cid, url)
	_, err2 := d.h.MapPut(DNSOwnerTable, url, owner)
	if err0 != nil || err1 != nil || err2 != nil {
		ilog.Errorf("write link mapPut failed. err = %v %v %v", err0, err1, err2)
	}
}

// RemoveLink remove a url
func (d *DNS) RemoveLink(url, cid string) {
	_, err0 := d.h.MapDel(DNSRTable, cid)
	_, err1 := d.h.MapDel(DNSTable, url)
	_, err2 := d.h.MapDel(DNSOwnerTable, url)
	if err0 != nil || err1 != nil || err2 != nil {
		ilog.Errorf("remove link mapPut failed. err = %v %v %v", err0, err1, err2)
	}
}
