package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetStoragePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetStoragePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyStorage)
	if err != nil {
		return
	}
	log.Printf("[%08X] get storage pool '%s' from %s.[%08X]", id, poolName, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)
	executor.ResourceModule.GetStoragePool(poolName, respChan)
	result := <-respChan
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetStoragePoolResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] get storage pool fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var poolInfo = result.StoragePool
	resp.SetString(vm_utils.ParamKeyName, poolInfo.Name)
	resp.SetString(vm_utils.ParamKeyType, poolInfo.Type)
	resp.SetString(vm_utils.ParamKeyHost, poolInfo.Host)
	resp.SetString(vm_utils.ParamKeyTarget, poolInfo.Target)
	resp.SetSuccess(true)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
