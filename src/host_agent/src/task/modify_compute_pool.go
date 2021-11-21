package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifyComputePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifyComputePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	pool, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	storagePool, _ := request.GetString(vm_utils.ParamKeyStorage)
	if "" == storagePool {
		log.Printf("[%08X] request modify compute pool '%s' using local storage from %s.[%08X]", id, pool, request.GetSender(), request.GetFromSession())
	} else {
		log.Printf("[%08X] request modify compute pool '%s' using storage pool '%s' from %s.[%08X]", id, pool, storagePool,
			request.GetSender(), request.GetFromSession())
	}
	addressPool, _ := request.GetString(vm_utils.ParamKeyNetwork)
	var failover = false
	failover, _ = request.GetBoolean(vm_utils.ParamKeyOption)

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyComputePoolResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if err = QualifyNormalName(pool); err != nil {
		log.Printf("[%08X] invalid pool name '%s' : %s", id, pool, err.Error())
		err = fmt.Errorf("invalid pool name '%s': %s", pool, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	var respChan = make(chan error)
	executor.ResourceModule.ModifyPool(pool, storagePool, addressPool, failover, respChan)
	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] modify compute pool fail: %s", id, err.Error())
	} else {
		resp.SetSuccess(true)
	}

	return executor.Sender.SendMessage(resp, request.GetSender())
}
