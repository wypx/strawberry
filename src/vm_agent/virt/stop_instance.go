package virt

import (
	"errors"
	"fmt"
	"log"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type StopInstanceExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *StopInstanceExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	instanceID, err = request.GetString(VmUtils.ParamKeyInstance)
	if err != nil {
		return err
	}
	options, err := request.GetUIntArray(VmUtils.ParamKeyOption)
	if err != nil {
		return err
	}
	const (
		ValidOptionCount = 2
	)
	if len(options) != ValidOptionCount {
		return fmt.Errorf("unexpected option count %d / %d", len(options), ValidOptionCount)
	}
	var reboot = 1 == options[0]
	var force = 1 == options[1]
	if reboot {
		if force {
			log.Printf("[%08X] request force reboot instance '%s' from %s.[%08X]",
				id, instanceID, request.GetSender(), request.GetFromSession())
		} else {
			log.Printf("[%08X] request reboot instance '%s' from %s.[%08X]",
				id, instanceID, request.GetSender(), request.GetFromSession())
		}
	} else if force {
		log.Printf("[%08X] request force stop instance '%s' from %s.[%08X]",
			id, instanceID, request.GetSender(), request.GetFromSession())
	} else {
		log.Printf("[%08X] request stop instance '%s' from %s.[%08X]",
			id, instanceID, request.GetSender(), request.GetFromSession())
	}

	var respChan = make(chan error)
	executor.InstanceModule.StopInstance(instanceID, reboot, force, respChan)
	err = <-respChan

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.StopInstanceResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	if err != nil {
		resp.SetSuccess(false)
		resp.SetError(err.Error())
		log.Printf("[%08X] stop instance fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] stop instance success", id)
	if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: send response fail: %s", id, err.Error())
		return err
	}
	if reboot {
		return nil
	}

	{
		//wait for instance stopped
		const (
			CheckInterval = 1 * time.Second
			WaitTimeout   = 1 * time.Minute
		)

		ticker := time.NewTicker(CheckInterval)
		timer := time.NewTimer(WaitTimeout)
		for {
			select {
			case <-ticker.C:
				var respChan = make(chan bool)
				executor.InstanceModule.IsInstanceRunning(instanceID, respChan)
				running := <-respChan
				if !running {
					log.Printf("[%08X] instance '%s' stopped", id, instanceID)
					event, _ := VmUtils.CreateJsonMessage(VmUtils.GuestStoppedEvent)
					event.SetFromSession(id)
					event.SetString(VmUtils.ParamKeyInstance, instanceID)
					if err = executor.Sender.SendMessage(event, request.GetSender()); err != nil {
						log.Printf("[%08X] warning: notify instance stopped fail: %s", id, err.Error())
					}
					return err
				}
			case <-timer.C:
				//timeout
				log.Printf("[%08X] warning: instance not stopped in expected duration", id)
				return errors.New("stop not finished")
			}
		}
	}
}
