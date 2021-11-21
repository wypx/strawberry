package task

import (
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryStoragePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryStoragePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	//log.Printf("[%08X] query storage pool from %s.[%08X]", id, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)
	executor.ResourceModule.QueryStoragePool(respChan)
	result := <-respChan
	var nameArray, typeArray, hostArray, targetArray []string
	for _, info := range result.StoragePoolList {
		nameArray = append(nameArray, info.Name)
		typeArray = append(typeArray, info.Type)
		hostArray = append(hostArray, info.Host)
		targetArray = append(targetArray, info.Target)
	}
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryStoragePoolResponse)
	resp.SetSuccess(true)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetStringArray(vm_utils.ParamKeyName, nameArray)
	resp.SetStringArray(vm_utils.ParamKeyType, typeArray)
	resp.SetStringArray(vm_utils.ParamKeyHost, hostArray)
	resp.SetStringArray(vm_utils.ParamKeyTarget, targetArray)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
