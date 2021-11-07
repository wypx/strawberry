package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type InsertMediaCoreExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *InsertMediaCoreExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID, mediaSource, host string
	var port uint
	instanceID, err = request.GetString(VmUtils.ParamKeyInstance)
	if err != nil {
		return err
	}
	if mediaSource, err = request.GetString(VmUtils.ParamKeyMedia); err != nil {
		return err
	}
	if host, err = request.GetString(VmUtils.ParamKeyHost); err != nil {
		return err
	}
	if port, err = request.GetUInt(VmUtils.ParamKeyPort); err != nil {
		return err
	}

	log.Printf("[%08X] request insert media '%s' into '%s' from %s.[%08X]", id, mediaSource, instanceID,
		request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.InsertMediaResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

	var respChan = make(chan error, 1)
	var media = VmAgentSvc.InstanceMediaConfig{Mode: VmAgentSvc.MediaModeHTTPS, ID: mediaSource, Host: host, Port: port}
	executor.InstanceModule.AttachMedia(instanceID, media, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] insert media fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] instance media inserted", id)
		resp.SetSuccess(true)
		{
			//notify event
			event, _ := VmUtils.CreateJsonMessage(VmUtils.MediaAttachedEvent)
			event.SetFromSession(id)
			event.SetString(VmUtils.ParamKeyInstance, instanceID)
			event.SetString(VmUtils.ParamKeyMedia, mediaSource)
			executor.Sender.SendMessage(event, request.GetSender())
		}
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
