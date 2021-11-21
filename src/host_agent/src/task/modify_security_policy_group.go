package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifySecurityPolicyGroupExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifySecurityPolicyGroupExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var policyID string
	if policyID, err = request.GetString(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy group ID fail: %s", err.Error())
		return
	}
	var config modules.SecurityPolicyGroup
	if config.Name, err = request.GetString(vm_utils.ParamKeyName); err != nil {
		err = fmt.Errorf("get name fail: %s", err.Error())
		return
	}
	if config.Description, err = request.GetString(vm_utils.ParamKeyDescription); err != nil {
		err = fmt.Errorf("get description fail: %s", err.Error())
		return
	}
	if config.User, err = request.GetString(vm_utils.ParamKeyUser); err != nil {
		err = fmt.Errorf("get user fail: %s", err.Error())
		return
	}
	if config.Group, err = request.GetString(vm_utils.ParamKeyGroup); err != nil {
		err = fmt.Errorf("get group fail: %s", err.Error())
		return
	}
	if config.Accept, err = request.GetBoolean(vm_utils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get accpet flag fail: %s", err.Error())
		return
	}
	if config.Enabled, err = request.GetBoolean(vm_utils.ParamKeyEnable); err != nil {
		err = fmt.Errorf("get enabled flag fail: %s", err.Error())
		return
	}
	if config.Global, err = request.GetBoolean(vm_utils.ParamKeyLimit); err != nil {
		err = fmt.Errorf("get global flag fail: %s", err.Error())
		return
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyPolicyGroupResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.ResourceModule.ModifySecurityPolicyGroup(policyID, config, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] modify security policy group '%s' fail: %s",
			id, policyID, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] security policy group '%s' modified",
			id, policyID)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
