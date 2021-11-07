package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type RemoveSecurityRuleExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *RemoveSecurityRuleExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	var index int
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance id fail: %s", err.Error())
		return
	}
	if index, err = request.GetInt(VmUtils.ParamKeyIndex); err != nil {
		err = fmt.Errorf("get rule index fail: %s", err.Error())
		return
	}
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.RemoveGuestRuleResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.InstanceModule.RemoveSecurityPolicyRule(instanceID, index, respChan)
	err = <-respChan
	if nil != err {
		log.Printf("[%08X] remove %dth security rule of instance '%s' fail: %s",
			id, index, instanceID, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] %dth security rule of instance '%s' removed",
			id, index, instanceID)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
