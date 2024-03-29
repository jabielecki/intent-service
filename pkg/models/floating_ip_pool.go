package models

import "github.com/tungstenfabric-preview/intent-service/pkg/common"

// IsParentTypeVirtualNetwork checks if parent's type is virtual network
func (fipp *FloatingIPPool) IsParentTypeVirtualNetwork() bool {
	var m VirtualNetwork
	return fipp.GetParentType() == m.Kind()
}

// HasSubnets checks if floating-ip-pool has any subnets defined
func (fipp *FloatingIPPool) HasSubnets() bool {
	floatingIPPoolSubnets := fipp.GetFloatingIPPoolSubnets()
	return floatingIPPoolSubnets != nil && len(floatingIPPoolSubnets.GetSubnetUUID()) != 0
}

// CheckAreSubnetsInVirtualNetworkSubnets checks if subnets defined in floating-ip-pool object
// are present in the virtual-network
func (fipp *FloatingIPPool) CheckAreSubnetsInVirtualNetworkSubnets(vn *VirtualNetwork) error {
	for _, floatingIPPoolSubnetUUID := range fipp.GetFloatingIPPoolSubnets().GetSubnetUUID() {
		subnetFound := false
		for _, ipam := range vn.GetNetworkIpamRefs() {
			for _, ipamSubnet := range ipam.GetAttr().GetIpamSubnets() {
				if ipamSubnet.GetSubnetUUID() == floatingIPPoolSubnetUUID {
					subnetFound = true
					break
				}
			}
		}

		if !subnetFound {
			return common.ErrorBadRequestf("Subnet %s was not found in virtual-network %s",
				floatingIPPoolSubnetUUID, vn.GetUUID())
		}
	}
	return nil
}
