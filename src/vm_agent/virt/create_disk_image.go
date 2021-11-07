package virt

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type CreateDiskImageExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
	Client         *http.Client
}

func (executor *CreateDiskImageExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var imageID, guestID, mediaHost string
	var mediaPort uint
	if imageID, err = request.GetString(VmUtils.ParamKeyImage); err != nil {
		return err
	}
	if guestID, err = request.GetString(VmUtils.ParamKeyGuest); err != nil {
		return err
	}
	if mediaHost, err = request.GetString(VmUtils.ParamKeyHost); err != nil {
		return err
	}
	if mediaPort, err = request.GetUInt(VmUtils.ParamKeyPort); err != nil {
		return err
	}
	log.Printf("[%08X] recv create disk image from %s.[%08X], from guest '%s' to image %s@%s:%d",
		id, request.GetSender(), request.GetFromSession(), guestID, imageID, mediaHost, mediaPort)
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.CreateDiskImageResponse)
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
		if !result.Instance.Created {
			err = fmt.Errorf("instance '%s' not created", guestID)
			log.Printf("[%08X] check guest status fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		if result.Instance.Running {
			err = fmt.Errorf("instance '%s' not stopped", guestID)
			log.Printf("[%08X] check guest status fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		if 0 == len(result.Instance.StorageVolumes) {
			err = errors.New("no volume available")
			log.Printf("[%08X] check guest status fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		targetVolume = result.Instance.StorageVolumes[0]
	}
	var startChan = make(chan error, 1)
	var progressChan = make(chan uint, 1)
	var resultChan = make(chan VmAgentSvc.StorageResult, 1)
	const (
		CheckInterval = 2 * time.Second
	)
	{

		//start write
		executor.StorageModule.WriteDiskImage(id, guestID, targetVolume, imageID, mediaHost, mediaPort, startChan, progressChan, resultChan)
		var timer = time.NewTimer(VmAgentSvc.DefaultOperateTimeout)
		select {
		case <-timer.C:
			err = errors.New("start write timeout")
			log.Printf("[%08X] write disk image fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		case err = <-startChan:
			if err != nil {
				log.Printf("[%08X] write disk image fail: %s", id, err.Error())
				resp.SetError(err.Error())
				return executor.Sender.SendMessage(resp, request.GetSender())
			}
			//start success
			log.Printf("[%08X] write disk image started", id)
			resp.SetSuccess(true)
			if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
				log.Printf("[%08X] warning: notify create start to '%s' fail: %s", id, request.GetSender(), err.Error())
			}

		}
	}
	event, _ := VmUtils.CreateJsonMessage(VmUtils.DiskImageUpdatedEvent)
	event.SetSuccess(true)
	event.SetFromSession(id)
	event.SetToSession(request.GetFromSession())

	{
		//wait progress & result
		var latestUpdate = time.Now()
		var ticker = time.NewTicker(CheckInterval)
		for {
			select {
			case <-ticker.C:
				//check
				if time.Now().After(latestUpdate.Add(VmAgentSvc.DefaultOperateTimeout)) {
					//timeout
					err = errors.New("timeout")
					log.Printf("[%08X] create disk image fail: %s", id, err.Error())
					event.SetSuccess(false)
					event.SetError(err.Error())
					return executor.Sender.SendMessage(event, request.GetSender())
				}
			case progress := <-progressChan:
				latestUpdate = time.Now()
				event.SetUInt(VmUtils.ParamKeyProgress, progress)
				event.SetBoolean(VmUtils.ParamKeyEnable, false)
				log.Printf("[%08X] progress => %d %%", id, progress)
				if err = executor.Sender.SendMessage(event, request.GetSender()); err != nil {
					log.Printf("[%08X] warning: notify progress fail: %s", id, err.Error())
				}
			case result := <-resultChan:
				err = result.Error
				if err != nil {
					log.Printf("[%08X] create disk image fail: %s", id, err.Error())
					event.SetSuccess(false)
					event.SetError(err.Error())
					return executor.Sender.SendMessage(event, request.GetSender())
				}
				log.Printf("[%08X] disk image written success, %d MB in size", id, result.Size>>20)
				event.SetBoolean(VmUtils.ParamKeyEnable, true)
				event.SetUInt(VmUtils.ParamKeySize, result.Size)
				event.SetUInt(VmUtils.ParamKeyProgress, 0)
				return executor.Sender.SendMessage(event, request.GetSender())
			}
		}
	}
}
