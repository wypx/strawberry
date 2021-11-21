package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryAddressRangeExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryAddressRangeExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var poolName, rangeType string
	if poolName, err = request.GetString(vm_utils.ParamKeyAddress); err != nil {
		return
	}
	if rangeType, err = request.GetString(vm_utils.ParamKeyType); err != nil {
		return
	}
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.QueryAddressRange(poolName, rangeType, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryAddressRangeResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] query address range from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var startArray, endArray, maskArray []string
	for _, status := range result.AddressRangeList {
		startArray = append(startArray, status.Start)
		endArray = append(endArray, status.End)
		maskArray = append(maskArray, status.Netmask)
	}
	resp.SetSuccess(true)
	resp.SetStringArray(vm_utils.ParamKeyStart, startArray)
	resp.SetStringArray(vm_utils.ParamKeyEnd, endArray)
	resp.SetStringArray(vm_utils.ParamKeyMask, maskArray)
	log.Printf("[%08X] reply %d address range(s) to %s.[%08X]",
		id, len(result.AddressRangeList), request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
