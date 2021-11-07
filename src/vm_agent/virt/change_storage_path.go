package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ChangeStoragePathExecutor struct {
	Sender  VmUtils.MessageSender
	Storage VmAgentSvc.StorageModule
}

func (executor *ChangeStoragePathExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var newPath string
	if newPath, err = request.GetString(VmUtils.ParamKeyPath); err != nil {
		err = fmt.Errorf("get new path fail: %s", err.Error())
		return
	}
	var respChan = make(chan error, 1)
	executor.Storage.ChangeDefaultStoragePath(newPath, respChan)

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyCellStorageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] change storage path fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	} else {
		resp.SetSuccess(true)
		log.Printf("[%08X] default storage path changed to: %s", id, newPath)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
