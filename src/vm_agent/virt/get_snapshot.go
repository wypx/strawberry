package virt

import (
	"errors"
	"log"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type GetSnapshotExecutor struct {
	Sender        VmUtils.MessageSender
	StorageModule VmAgentSvc.StorageModule
}

func (executor *GetSnapshotExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID, snapshotName string
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		return err
	}
	if snapshotName, err = request.GetString(VmUtils.ParamKeyName); err != nil {
		return err
	}

	log.Printf("[%08X] recv get snapshot '%s' for guest '%s' from %s.[%08X]",
		id, snapshotName, instanceID, request.GetSender(), request.GetFromSession())
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.GetSnapshotResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	{
		var respChan = make(chan VmAgentSvc.StorageResult, 1)
		executor.StorageModule.GetSnapshot(instanceID, snapshotName, respChan)
		var timer = time.NewTimer(VmAgentSvc.DefaultOperateTimeout)
		select {
		case <-timer.C:
			err = errors.New("request timeout")
			log.Printf("[%08X] get snapshot timeout", id)
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		case result := <-respChan:
			if result.Error != nil {
				err = result.Error
				log.Printf("[%08X] get snapshot fail: %s", id, err.Error())
				resp.SetError(err.Error())
			} else {
				var snapshot = result.Snapshot
				resp.SetBoolean(VmUtils.ParamKeyStatus, snapshot.Running)
				resp.SetString(VmUtils.ParamKeyDescription, snapshot.Description)
				resp.SetString(VmUtils.ParamKeyCreate, snapshot.CreateTime)
				resp.SetSuccess(true)
			}
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
