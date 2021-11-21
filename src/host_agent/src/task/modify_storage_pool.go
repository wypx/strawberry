package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifyStoragePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifyStoragePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var pool, storageType, host, target string
	pool, err = request.GetString(vm_utils.ParamKeyStorage)
	if err != nil {
		return
	}
	if storageType, err = request.GetString(vm_utils.ParamKeyType); err != nil {
		return
	}
	if host, err = request.GetString(vm_utils.ParamKeyHost); err != nil {
		return
	}
	if target, err = request.GetString(vm_utils.ParamKeyTarget); err != nil {
		return
	}

	log.Printf("[%08X] request modify storage pool '%s' from %s.[%08X]", id, pool,
		request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyStoragePoolResponse)
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
	executor.ResourceModule.ModifyStoragePool(pool, storageType, host, target, respChan)
	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] modify storage pool fail: %s", id, err.Error())
	} else {
		resp.SetSuccess(true)
	}

	return executor.Sender.SendMessage(resp, request.GetSender())
}
