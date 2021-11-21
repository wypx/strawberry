package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryAddressPoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryAddressPoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.QueryAddressPool(respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryAddressPoolResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] query address pool from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var nameArray, gatewayArray, dnsArray, providerArray []string
	var addressArray, allocateArray, dnsCountArray []uint64
	for _, pool := range result.AddressPoolList {
		nameArray = append(nameArray, pool.Name)
		gatewayArray = append(gatewayArray, pool.Gateway)
		providerArray = append(providerArray, pool.Provider)
		var addressCount uint32 = 0
		allocateArray = append(allocateArray, uint64(len(pool.Allocated)))
		for _, addressRange := range pool.Ranges {
			addressCount += addressRange.Capacity
		}
		addressArray = append(addressArray, uint64(addressCount))
		dnsCountArray = append(dnsCountArray, uint64(len(pool.DNS)))
		dnsArray = append(dnsArray, pool.DNS...)
	}
	resp.SetSuccess(true)
	resp.SetStringArray(vm_utils.ParamKeyName, nameArray)
	resp.SetStringArray(vm_utils.ParamKeyGateway, gatewayArray)
	resp.SetStringArray(vm_utils.ParamKeyServer, dnsArray)
	resp.SetStringArray(vm_utils.ParamKeyMode, providerArray)
	resp.SetUIntArray(vm_utils.ParamKeyAddress, addressArray)
	resp.SetUIntArray(vm_utils.ParamKeyAllocate, allocateArray)
	resp.SetUIntArray(vm_utils.ParamKeyCount, dnsCountArray)
	log.Printf("[%08X] reply %d address pool(s) to %s.[%08X]",
		id, len(result.AddressPoolList), request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
