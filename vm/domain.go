package vm

import (
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

// const ...
const (
	DHCPTable  = "dhcp_table"
	DHCPRTable = "dhcp_revert_table"
)

// DHCP ...
type DHCP struct {
	h *host.Host
}

// NewDHCP ...
func NewDHCP(h *host.Host) *DHCP {
	return &DHCP{
		h: h,
	}
}

// ContractID .
func (d *DHCP) ContractID(url string) string {
	cid, _ := d.h.MapGet(DHCPTable, url)
	if s, ok := cid.(string); ok {
		return s
	}
	return ""
}

// URL .
func (d *DHCP) URL(cid string) string {
	domain, _ := d.h.MapGet(DHCPRTable, cid)
	if s, ok := domain.(string); ok {
		return s
	}
	return ""
}

// IsDomain .
func (d *DHCP) IsDomain(s string) bool {
	if strings.HasPrefix(s, "Contract") || strings.HasPrefix(s, "iost.") {
		return false
	}
	return true
}

func (d *DHCP) Write(url, cid string) {
	d.h.MapPut(DHCPTable, url, cid)
	d.h.MapPut(DHCPRTable, cid, url)
}

func (d *DHCP) Remove(url, cid string) {
	d.h.MapDel(DHCPRTable, cid)
	d.h.MapDel(DHCPTable, url)
}
