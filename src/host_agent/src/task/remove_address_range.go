package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type RemoveAddressRangeExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *RemoveAddressRangeExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
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
	var respChan = make(chan error, 1)
	executor.ResourceModule.RemoveAddressRange(poolName, rangeType, startAddress, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.RemoveAddressRangeResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] request remove address range from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] range '%s' removed from pool '%s' by %s.[%08X]",
		id, startAddress,
		poolName, request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
