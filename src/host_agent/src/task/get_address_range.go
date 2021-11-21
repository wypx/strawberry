package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetAddressRangeExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetAddressRangeExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {

	var poolName, rangeType, startAddress string
	if poolName, err = request.GetString(vm_utils.ParamKeyAddress); err != nil {
		return
	}
	if rangeType, err = request.GetString(vm_utils.ParamKeyType); err != nil {
		return
	}
	if startAddress, err = request.GetString(vm_utils.ParamKeyStart); err != nil {
		return
	}
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetAddressRange(poolName, rangeType, startAddress, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetAddressRangeResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] request get address range from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var status = result.AddressRange

	var addressArray, instanceArray []string
	for _, allocated := range status.Allocated {
		addressArray = append(addressArray, allocated.Address)
		instanceArray = append(instanceArray, allocated.Instance)
	}

	resp.SetSuccess(true)
	resp.SetString(vm_utils.ParamKeyStart, status.Start)
	resp.SetString(vm_utils.ParamKeyEnd, status.End)
	resp.SetString(vm_utils.ParamKeyMask, status.Netmask)
	resp.SetUInt(vm_utils.ParamKeyCount, uint(status.Capacity))
	resp.SetStringArray(vm_utils.ParamKeyAddress, addressArray)
	resp.SetStringArray(vm_utils.ParamKeyInstance, instanceArray)
	log.Printf("[%08X] reply status of address range '%s' to %s.[%08X]",
		id, startAddress, request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
