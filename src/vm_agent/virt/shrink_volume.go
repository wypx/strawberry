package virt

import (
	"errors"
	"fmt"
	"log"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ShrinkGuestVolumeExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
}

func (executor *ShrinkGuestVolumeExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var guestID string
	var index uint
	if guestID, err = request.GetString(VmUtils.ParamKeyGuest); err != nil {
		return err
	}
	if index, err = request.GetUInt(VmUtils.ParamKeyDisk); err != nil {
		return err
	}
	log.Printf("[%08X] recv shrink disk of guest '%s' from %s.[%08X]",
		id, guestID, request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ResizeDiskResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	var targetVolume string
	{
		var respChan = make(chan VmAgentSvc.InstanceResult)
		executor.InstanceModule.GetInstanceStatus(guestID, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] get instance fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}

		err = func(instance VmAgentSvc.InstanceStatus, index int) (err error) {
			if !instance.Created {
				err = fmt.Errorf("instance '%s' not created", guestID)
				return
			}
			if instance.Running {
				err = fmt.Errorf("instance '%s' not stopped", guestID)
				return
			}
			var volumeCount = len(instance.StorageVolumes)
			if 0 == volumeCount {
				err = errors.New("no volume available")
				return
			}
			if index >= volumeCount {
				err = fmt.Errorf("invalid disk index %d", index)
				return
			}
			return nil
		}(result.Instance, int(index))
		if err != nil {
			log.Printf("[%08X] check instance fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		targetVolume = result.Instance.StorageVolumes[int(index)]
	}
	var resultChan = make(chan VmAgentSvc.StorageResult, 1)
	{
		executor.StorageModule.ShrinkVolume(id, guestID, targetVolume, resultChan)
		var timer = time.NewTimer(VmAgentSvc.DefaultOperateTimeout)
		select {
		case <-timer.C:
			err = errors.New("request timeout")
			log.Printf("[%08X] shrink disk timeout", id)
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		case result := <-resultChan:
			if result.Error != nil {
				err = result.Error
				log.Printf("[%08X] shrink disk fail: %s", id, err.Error())
				resp.SetError(err.Error())
			} else {
				log.Printf("[%08X] volume %s shrank successfully", id, targetVolume)
				resp.SetSuccess(true)
			}
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
