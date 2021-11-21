package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type DeleteStoragePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *DeleteStoragePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	pool, err := request.GetString(vm_utils.ParamKeyStorage)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request delete storage pool '%s' from %s.[%08X]", id, pool, request.GetSender(), request.GetFromSession())
	var respChan = make(chan error)

	executor.ResourceModule.DeleteStoragePool(pool, respChan)

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteStoragePoolResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] delete storage pool fail: %s", id, err.Error())
	} else {
		resp.SetSuccess(true)
	}

	return executor.Sender.SendMessage(resp, request.GetSender())
}
