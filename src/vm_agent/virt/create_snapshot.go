package virt

import (
	"errors"
	"fmt"
	"log"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type CreateSnapshotExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
}

func (executor *CreateSnapshotExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID, snapshot, description string
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		return err
	}
	if snapshot, err = request.GetString(VmUtils.ParamKeyName); err != nil {
		return err
	}
	description, _ = request.GetString(VmUtils.ParamKeyDescription)

	log.Printf("[%08X] recv create snapshot '%s' for guest '%s' from %s.[%08X]",
		id, snapshot, instanceID, request.GetSender(), request.GetFromSession())
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.CreateSnapshotResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	{
		var respChan = make(chan VmAgentSvc.InstanceResult, 1)
		executor.InstanceModule.GetInstanceStatus(instanceID, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] get instance fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}

		err = func(instance VmAgentSvc.InstanceStatus) (err error) {
			if !instance.Created {
				err = fmt.Errorf("instance '%s' not created", instanceID)
				return
			}
			//todo: allow operating on branch snapshots
			if instance.Running {
				err = errors.New("live snapshot not supported yes, shutdown instance first")
				return
			}
			return nil
		}(result.Instance)
		if err != nil {
			log.Printf("[%08X] check instance fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		var respChan = make(chan error, 1)
		executor.StorageModule.CreateSnapshot(instanceID, snapshot, description, respChan)
		var timer = time.NewTimer(VmAgentSvc.DefaultOperateTimeout)
		select {
		case <-timer.C:
			err = errors.New("request timeout")
			log.Printf("[%08X] create snapshot timeout", id)
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		case err = <-respChan:
			if err != nil {
				log.Printf("[%08X] create snapshot fail: %s", id, err.Error())
				resp.SetError(err.Error())
			} else {
				log.Printf("[%08X] snapshot '%s' created for guest '%s'", id, snapshot, instanceID)
				resp.SetSuccess(true)
			}
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
