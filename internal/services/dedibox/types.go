package dedibox

import (
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenCPUs(cpus []*dedibox.CPU) any {
	if cpus == nil {
		return nil
	}

	flattenedCPUs := []map[string]any(nil)
	for _, cpu := range cpus {
		flattenedCPUs = append(flattenedCPUs, map[string]any{
			"name":         cpu.Name,
			"core_count":   cpu.CoreCount,
			"frequency":    cpu.Frequency,
			"thread_count": cpu.ThreadCount,
		})
	}

	return flattenedCPUs
}

func flattenDisks(disks []*dedibox.Disk) any {
	if disks == nil {
		return nil
	}

	flattenedDisks := []map[string]any(nil)
	for _, disk := range disks {
		flattenedDisks = append(flattenedDisks, map[string]any{
			"type":     disk.Type,
			"capacity": disk.Capacity,
		})
	}

	return flattenedDisks
}

func flattenMemories(memories []*dedibox.Memory) any {
	if memories == nil {
		return nil
	}

	flattenedMemories := []map[string]any(nil)
	for _, memory := range memories {
		flattenedMemories = append(flattenedMemories, map[string]any{
			"type":      memory.Type,
			"capacity":  memory.Capacity,
			"frequency": memory.Frequency,
			"is_ecc":    memory.IsEcc,
		})
	}

	return flattenedMemories
}

func flattenNetworkInterfaces(interfaces []*dedibox.NetworkInterface) any {
	if interfaces == nil {
		return nil
	}

	flattenedInterfaces := []map[string]any(nil)
	for _, iface := range interfaces {
		flattenedInterfaces = append(flattenedInterfaces, map[string]any{
			"type": iface.Type,
			"mac":  iface.Mac,
			"ips":  flattenIPs(iface.IPs),
		})
	}

	return flattenedInterfaces
}

func flattenIPs(ips []*dedibox.IP) any {
	if ips == nil {
		return nil
	}

	flattenedIPs := []map[string]any(nil)
	for _, ip := range ips {
		flattenedIPs = append(flattenedIPs, map[string]any{
			"ip_id":    ip.IPID,
			"address":  ip.Address.String(),
			"reverse":  ip.Reverse,
			"version":  ip.Version.String(),
			"cidr":     ip.Cidr,
			"netmask":  ip.Netmask,
			"semantic": ip.Semantic.String(),
			"gateway":  ip.Gateway,
			"status":   ip.Status.String(),
		})
	}

	return flattenedIPs
}

func flattenServerOptions(options []*dedibox.ServerOption) any {
	if options == nil {
		return nil
	}

	flattenedOptions := []map[string]any(nil)
	for _, option := range options {
		optionMap := map[string]any{
			"expires_at": types.FlattenTime(option.ExpiredAt),
		}
		if option.Offer != nil {
			optionMap["id"] = option.Offer.ID
			optionMap["name"] = option.Offer.Name
		}
		flattenedOptions = append(flattenedOptions, optionMap)
	}

	return flattenedOptions
}

func flattenServerLocation(location *dedibox.ServerLocation) map[string]any {
	if location == nil {
		return nil
	}

	return map[string]any{
		"rack":       location.Rack,
		"room":       location.Room,
		"datacenter": location.DatacenterName,
	}
}

func flattenOS(os *dedibox.OS) map[string]any {
	if os == nil {
		return nil
	}

	return map[string]any{
		"id":      os.ID,
		"name":    os.Name,
		"version": os.Version,
	}
}

func flattenServiceLevel(level *dedibox.ServiceLevel) map[string]any {
	if level == nil {
		return nil
	}

	return map[string]any{
		"offer_id": level.OfferID,
		"level":    level.Level.String(),
	}
}
