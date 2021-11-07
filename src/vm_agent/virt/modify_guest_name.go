package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyGuestNameExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifyGuestNameExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) error {
	guestID, err := request.GetString(VmUtils.ParamKeyGuest)
	if err != nil {
		return err
	}
	name, err := request.GetString(VmUtils.ParamKeyName)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request rename guest '%s' from %s.[%08X]", id, guestID,
		request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyGuestNameResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var respChan = make(chan error)
	executor.InstanceModule.ModifyGuestName(guestID, name, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] rename guest fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] guest '%s' renamed to %s", id, guestID, name)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
