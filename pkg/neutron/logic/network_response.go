package logic

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/models"
)

func makeNetworkResponse(rp RequestParameters, vn *models.VirtualNetwork, oper string) *NetworkResponse {
	parentNeutronUUID := VncUUIDToNeutronID(vn.GetParentUUID())
	nn := &NetworkResponse{
		ID:                      vn.GetUUID(),
		Name:                    vn.GetDisplayName(),
		TenantID:                parentNeutronUUID,
		ProjectID:               parentNeutronUUID,
		AdminStateUp:            vn.GetIDPerms().GetEnable(),
		Shared:                  vn.GetIsShared(),
		Status:                  netStatusDown,
		RouterExternal:          vn.GetRouterExternal(),
		PortSecurityEnabled:     vn.GetPortSecurityEnabled(),
		Description:             vn.GetIDPerms().GetDescription(),
		CreatedAt:               vn.GetIDPerms().GetCreated(),
		UpdatedAt:               vn.GetIDPerms().GetLastModified(),
		ProviderPhysicalNetwork: vn.GetProviderProperties().GetPhysicalNetwork(),
		ProviderSegmentationID:  vn.GetProviderProperties().GetSegmentationID(),
		Subnets:                 []string{},
		SubnetIpam:              []*SubnetIpam{},
	}

	if contrailExtensionsEnabled {
		nn.FQName = vn.GetFQName()
	}

	if vn.GetDisplayName() == "" {
		nn.Name = vn.FQName[len(vn.FQName)-1]
	}

	if !nn.Shared && (vn.GetPerms2() != nil && isSharedWithTenant(&rp.RequestContext, vn.GetPerms2().GetShare())) {
		nn.Shared = true
	}

	if vn.GetIDPerms().GetEnable() {
		nn.Status = netStatusActive
	}

	if prop := vn.GetProviderProperties(); prop != nil {
		nn.ProviderPhysicalNetwork = prop.GetPhysicalNetwork()
		nn.ProviderSegmentationID = prop.GetSegmentationID()
	}

	if contrailExtensionsEnabled {
		nn.setResponseRefs(vn, oper)
	}

	nn.setSubnets(vn)

	return nn
}

func (r *NetworkResponse) setResponseRefs(vn *models.VirtualNetwork, oper string) {
	if oper == OperationRead || oper == OperationReadAll {
		r.setPolicys(vn)
	}
	r.setRouteTables(vn)
}

func (r *NetworkResponse) setPolicys(vn *models.VirtualNetwork) {
	for _, np := range vn.GetNetworkPolicyRefs() {
		r.Policys = append(r.Policys, np.GetTo())
	}
}

func (r *NetworkResponse) setRouteTables(vn *models.VirtualNetwork) {
	for _, rt := range vn.GetRouteTableRefs() {
		r.RouteTable = append(r.RouteTable, rt.GetTo())
	}
}

func (r *NetworkResponse) setSubnets(vn *models.VirtualNetwork) {
	ipamRefs := vn.GetNetworkIpamRefs()
	for _, ipam := range ipamRefs {
		subnets := ipam.GetAttr().GetIpamSubnets()
		for _, ipamSubnet := range subnets {
			sn := subnetVncToNeutron(vn, ipamSubnet)
			r.Subnets = append(r.Subnets, sn.ID)

			if contrailExtensionsEnabled {
				r.SubnetIpam = append(r.SubnetIpam, &SubnetIpam{SubnetCidr: sn.Cidr, IpamFQName: ipam.GetTo()})

			}
		}
	}
}
