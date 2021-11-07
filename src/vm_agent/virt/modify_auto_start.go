package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyAutoStartExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifyAutoStartExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var guestID string
	var enable bool
	if guestID, err = request.GetString(VmUtils.ParamKeyGuest); err != nil {
		err = fmt.Errorf("get guest id fail: %s", err.Error())
		return
	}
	if enable, err = request.GetBoolean(VmUtils.ParamKeyEnable); err != nil {
		err = fmt.Errorf("get enable flag fail: %s", err.Error())
		return
	}
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyAutoStartResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.InstanceModule.ModifyAutoStart(guestID, enable, respChan)
	if err = <-respChan; err != nil {
		log.Printf("[%08X] modify auto start fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		if enable {
			log.Printf("[%08X] auto start of guest '%s' enabled", id, guestID)
		} else {
			log.Printf("[%08X] auto start of guest '%s' disabled", id, guestID)
		}
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
