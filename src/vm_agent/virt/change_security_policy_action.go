package virt

import (
	"fmt"
	"log"

	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ChangeDefaultSecurityActionExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ChangeDefaultSecurityActionExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	var accept bool
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance id fail: %s", err.Error())
		return
	}
	if accept, err = request.GetBoolean(VmUtils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get action fail: %s", err.Error())
		return
	}
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ChangeGuestRuleDefaultActionResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.InstanceModule.ChangeDefaultSecurityPolicyAction(instanceID, accept, respChan)
	err = <-respChan
	if nil != err {
		log.Printf("[%08X] change default security policy action of instance '%s' fail: %s",
			id, instanceID, err.Error())
		resp.SetError(err.Error())
	} else {
		if accept {
			log.Printf("[%08X] default security policy action of instance '%s' changed to accept",
				id, instanceID)
		} else {
			log.Printf("[%08X] default security policy action of instance '%s' changed to drop",
				id, instanceID)
		}
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
