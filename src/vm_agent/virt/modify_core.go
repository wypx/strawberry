package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyGuestCoreExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifyGuestCoreExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) error {
	guestID, err := request.GetString(VmUtils.ParamKeyGuest)
	if err != nil {
		return err
	}
	cores, err := request.GetUInt(VmUtils.ParamKeyCore)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request modifying cores of '%s' from %s.[%08X]", id, guestID,
		request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyCoreResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var respChan = make(chan error)
	executor.InstanceModule.ModifyGuestCore(guestID, cores, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] modify core fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] cores of guest '%s' changed to %d", id, guestID, cores)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
