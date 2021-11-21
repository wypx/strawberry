package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetAddressPoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetAddressPoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var poolName string
	if poolName, err = request.GetString(vm_utils.ParamKeyAddress); err != nil {
		return
	}
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetAddressPool(poolName, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetAddressPoolResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] get address pool from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var status = result.AddressPool
	var startArray, endArray, maskArray []string
	var capacityArray []uint64
	for _, addressRange := range status.Ranges {
		startArray = append(startArray, addressRange.Start)
		endArray = append(endArray, addressRange.End)
		maskArray = append(maskArray, addressRange.Netmask)
		capacityArray = append(capacityArray, uint64(addressRange.Capacity))
	}

	var addressArray, instanceArray []string
	for _, allocated := range status.Allocated {
		addressArray = append(addressArray, allocated.Address)
		instanceArray = append(instanceArray, allocated.Instance)
	}
	resp.SetSuccess(true)
	resp.SetString(vm_utils.ParamKeyGateway, status.Gateway)
	resp.SetString(vm_utils.ParamKeyMode, status.Provider)
	resp.SetStringArray(vm_utils.ParamKeyServer, status.DNS)
	resp.SetStringArray(vm_utils.ParamKeyStart, startArray)
	resp.SetStringArray(vm_utils.ParamKeyEnd, endArray)
	resp.SetStringArray(vm_utils.ParamKeyMask, maskArray)
	resp.SetUIntArray(vm_utils.ParamKeyCount, capacityArray)
	resp.SetStringArray(vm_utils.ParamKeyAddress, addressArray)
	resp.SetStringArray(vm_utils.ParamKeyInstance, instanceArray)
	log.Printf("[%08X] reply status of address pool '%s' to %s.[%08X]",
		id, poolName, request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
