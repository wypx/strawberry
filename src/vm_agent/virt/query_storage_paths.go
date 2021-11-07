package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type QueryStoragePathExecutor struct {
	Sender  VmUtils.MessageSender
	Storage VmAgentSvc.StorageModule
}

func (executor *QueryStoragePathExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var respChan = make(chan VmAgentSvc.StorageResult, 1)
	executor.Storage.QueryStoragePaths(respChan)

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.QueryCellStorageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] query storage paths fail: %s", id, err.Error())
	} else {
		//parse result
		resp.SetSuccess(true)
		resp.SetUInt(VmUtils.ParamKeyMode, uint(result.StorageMode))
		resp.SetStringArray(VmUtils.ParamKeySystem, result.SystemPaths)
		resp.SetStringArray(VmUtils.ParamKeyData, result.DataPaths)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
