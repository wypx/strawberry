package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type EjectMediaCoreExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *EjectMediaCoreExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) error {
	instanceID, err := request.GetString(VmUtils.ParamKeyInstance)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request eject media from '%s' from %s.[%08X]", id, instanceID,
		request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.EjectMediaResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

	var respChan = make(chan error, 1)
	executor.InstanceModule.DetachMedia(instanceID, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] eject media fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] instance media ejected", id)
		resp.SetSuccess(true)
		{
			//notify event
			event, _ := VmUtils.CreateJsonMessage(VmUtils.MediaDetachedEvent)
			event.SetFromSession(id)
			event.SetString(VmUtils.ParamKeyInstance, instanceID)
			executor.Sender.SendMessage(event, request.GetSender())
		}
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
