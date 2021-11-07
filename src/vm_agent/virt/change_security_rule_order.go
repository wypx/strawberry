package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ChangeSecurityRuleOrderExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ChangeSecurityRuleOrderExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	var direction, index int
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance id fail: %s", err.Error())
		return
	}
	if index, err = request.GetInt(VmUtils.ParamKeyIndex); err != nil {
		err = fmt.Errorf("get index fail: %s", err.Error())
		return
	}
	if direction, err = request.GetInt(VmUtils.ParamKeyMode); err != nil {
		err = fmt.Errorf("get direction fail: %s", err.Error())
		return
	}
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ChangeGuestRuleOrderResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	var moveUp = false
	if direction >= 0 {
		moveUp = true
		executor.InstanceModule.PullUpSecurityPolicyRule(instanceID, index, respChan)
	} else {
		executor.InstanceModule.PushDownSecurityPolicyRule(instanceID, index, respChan)
	}

	err = <-respChan
	if nil != err {
		log.Printf("[%08X] change order of %dth security rule of instance '%s' fail: %s",
			id, index, instanceID, err.Error())
		resp.SetError(err.Error())
	} else {
		if moveUp {
			log.Printf("[%08X] %dth security rule of instance '%s' moved up",
				id, index, instanceID)
		} else {
			log.Printf("[%08X] %dth security rule of instance '%s' moved down",
				id, index, instanceID)
		}
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
