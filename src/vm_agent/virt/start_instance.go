package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type StartInstanceExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *StartInstanceExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	const (
		InstanceMediaOptionNone uint = iota
		InstanceMediaOptionImage
		InstanceMediaOptionNetwork
	)

	var instanceID string
	instanceID, err = request.GetString(VmUtils.ParamKeyInstance)
	if err != nil {
		return err
	}
	mediaOption, err := request.GetUInt(VmUtils.ParamKeyOption)
	if err != nil {
		return err
	}
	var respChan = make(chan error)
	var mediaSource string

	switch mediaOption {
	case InstanceMediaOptionNone:
		//no media attached
		log.Printf("[%08X] request start instance '%s' from %s.[%08X]",
			id, instanceID, request.GetSender(), request.GetFromSession())
		executor.InstanceModule.StartInstance(instanceID, respChan)
	case InstanceMediaOptionImage:
		var host string
		var port uint
		if host, err = request.GetString(VmUtils.ParamKeyHost); err != nil {
			return err
		}
		if mediaSource, err = request.GetString(VmUtils.ParamKeySource); err != nil {
			return err
		}
		if port, err = request.GetUInt(VmUtils.ParamKeyPort); err != nil {
			return err
		}
		var media = VmAgentSvc.InstanceMediaConfig{Mode: VmAgentSvc.MediaModeHTTPS, ID: mediaSource, Host: host, Port: port}
		log.Printf("[%08X] request start instance '%s' with media '%s' (host %s:%d) from %s.[%08X]",
			id, instanceID, mediaSource, host, port, request.GetSender(), request.GetFromSession())
		executor.InstanceModule.StartInstanceWithMedia(instanceID, media, respChan)
	default:
		return fmt.Errorf("unsupported media option %d", mediaOption)
	}

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.StartInstanceResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	err = <-respChan
	if err != nil {
		resp.SetSuccess(false)
		resp.SetError(err.Error())
		log.Printf("[%08X] start instance fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] start instance success", id)
	if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: send response fail: %s", id, err.Error())
		return err
	}
	//notify
	event, _ := VmUtils.CreateJsonMessage(VmUtils.GuestStartedEvent)
	event.SetFromSession(id)
	event.SetString(VmUtils.ParamKeyInstance, instanceID)
	if err = executor.Sender.SendMessage(event, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: notify instance started fail: %s", id, err.Error())
		return err
	}
	if InstanceMediaOptionImage == mediaOption {
		//notify media attached
		attached, _ := VmUtils.CreateJsonMessage(VmUtils.MediaAttachedEvent)
		attached.SetFromSession(id)
		attached.SetString(VmUtils.ParamKeyInstance, instanceID)
		attached.SetString(VmUtils.ParamKeyMedia, mediaSource)
		if err = executor.Sender.SendMessage(attached, request.GetSender()); err != nil {
			log.Printf("[%08X] warning: notify media attached fail: %s", id, err.Error())
			return err
		}
	}
	return nil
}
