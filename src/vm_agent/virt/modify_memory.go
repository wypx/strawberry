package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyGuestMemoryExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifyGuestMemoryExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) error {
	guestID, err := request.GetString(VmUtils.ParamKeyGuest)
	if err != nil {
		return err
	}
	memory, err := request.GetUInt(VmUtils.ParamKeyMemory)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request modifying memory of '%s' from %s.[%08X]", id, guestID,
		request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyMemoryResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var respChan = make(chan error)
	executor.InstanceModule.ModifyGuestMemory(guestID, memory, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] modify memory fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] memory of guest '%s' changed to %d MB", id, guestID, memory/(1<<20))
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
