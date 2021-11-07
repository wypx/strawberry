package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type GetInstanceConfigExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *GetInstanceConfigExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	instanceID, err = request.GetString(VmUtils.ParamKeyInstance)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request get config of instance '%s' from %s.[%08X]",
		id, instanceID, request.GetSender(), request.GetFromSession())
	var respChan = make(chan VmAgentSvc.InstanceResult)
	executor.InstanceModule.GetInstanceConfig(instanceID, respChan)

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.GetGuestResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	result := <-respChan
	if result.Error != nil {
		resp.SetSuccess(false)
		resp.SetError(result.Error.Error())
		log.Printf("[%08X] get instance status fail: %s", id, result.Error.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var c = result.Instance.GuestConfig
	resp.SetSuccess(true)
	c.Marshal(resp)

	log.Printf("[%08X] query instance config success", id)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
