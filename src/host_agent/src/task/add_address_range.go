package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type AddAddressRangeExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *AddAddressRangeExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var poolName, rangeType string
	if poolName, err = request.GetString(vm_utils.ParamKeyAddress); err != nil {
		return
	}
	if rangeType, err = request.GetString(vm_utils.ParamKeyType); err != nil {
		return
	}
	var config modules.AddressRangeConfig
	if config.Start, err = request.GetString(vm_utils.ParamKeyStart); err != nil {
		return
	}
	if config.End, err = request.GetString(vm_utils.ParamKeyEnd); err != nil {
		return
	}
	if config.Netmask, err = request.GetString(vm_utils.ParamKeyMask); err != nil {
		return
	}
	var respChan = make(chan error, 1)
	executor.ResourceModule.AddAddressRange(poolName, rangeType, config, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.AddAddressRangeResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] request add address range from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] range '%s ~ %s/%s' added to pool '%s' from %s.[%08X]",
		id, config.Start, config.End, config.Netmask,
		poolName, request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
