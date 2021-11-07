package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyCPUPriorityExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifyCPUPriorityExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) error {
	guestID, err := request.GetString(VmUtils.ParamKeyGuest)
	if err != nil {
		return err
	}
	priorityValue, err := request.GetUInt(VmUtils.ParamKeyPriority)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request changing CPU priority of guest '%s' to %d from %s.[%08X]", id, guestID,
		priorityValue, request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyPriorityResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.InstanceModule.ModifyCPUPriority(guestID, VmAgentSvc.PriorityEnum(priorityValue), respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] modify CPU priority fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] CPU priority of guest '%s' changed to %d", id, guestID, priorityValue)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
