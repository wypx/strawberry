package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type CreateAddressPoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *CreateAddressPoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var config modules.AddressPoolConfig
	var poolName string
	if poolName, err = request.GetString(vm_utils.ParamKeyAddress); err != nil {
		return
	}
	config.Name = poolName
	if config.Gateway, err = request.GetString(vm_utils.ParamKeyGateway); err != nil {
		return
	}
	if config.DNS, err = request.GetStringArray(vm_utils.ParamKeyServer); err != nil {
		return
	}
	if config.Provider, err = request.GetString(vm_utils.ParamKeyMode); err != nil {
		err = fmt.Errorf("get provider fail: %s", err.Error())
		return
	}
	var respChan = make(chan error, 1)
	executor.ResourceModule.CreateAddressPool(config, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.CreateAddressPoolResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] request create address pool from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	resp.SetSuccess(true)
	log.Printf("[%08X] address pool '%s' created from %s.[%08X]",
		id, poolName, request.GetSender(), request.GetFromSession())
	return executor.Sender.SendMessage(resp, request.GetSender())
}
