package virt

import (
	"fmt"
	"log"
	"strings"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type HandleAddressPoolChangedExecutor struct {
	InstanceModule VmAgentSvc.InstanceModule
	NetworkModule  VmAgentSvc.NetworkModule
}

func (executor *HandleAddressPoolChangedExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var allocationMode, gateway string
	var dns []string
	if gateway, err = request.GetString(VmUtils.ParamKeyGateway); err != nil {
		return err
	}
	if dns, err = request.GetStringArray(VmUtils.ParamKeyServer); err != nil {
		return
	}
	if allocationMode, err = request.GetString(VmUtils.ParamKeyMode); err != nil {
		err = fmt.Errorf("get allocation mode fail: %s", err.Error())
		return
	}
	switch allocationMode {
	case VmAgentSvc.AddressAllocationNone:
	case VmAgentSvc.AddressAllocationDHCP:
	case VmAgentSvc.AddressAllocationCloudInit:
		break
	default:
		err = fmt.Errorf("invalid allocation mode :%s", allocationMode)
		return
	}
	var respChan = make(chan error, 1)
	executor.NetworkModule.UpdateAddressAllocation(gateway, dns, allocationMode, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] update address allocation fail when address pool changed from %s.[%08X]: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
	} else {
		log.Printf("[%08X] address allocation updated to mode %s, gateway: %s, DNS: %s",
			id, allocationMode, gateway, strings.Join(dns, "/"))
		if VmAgentSvc.AddressAllocationNone != allocationMode {
			executor.InstanceModule.SyncAddressAllocation(allocationMode)
		}
	}
	return nil
}
