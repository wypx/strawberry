package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type CreateSecurityPolicyGroupExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *CreateSecurityPolicyGroupExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
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

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.CreatePolicyGroupResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.CreateSecurityPolicyGroup(config, respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] create security policy group '%s' fail: %s",
			id, config.Name, err.Error())
		resp.SetError(err.Error())
	} else {
		var policy = result.PolicyGroup
		log.Printf("[%08X] new security policy group '%s'('%s') created",
			id, config.Name, policy.ID)
		resp.SetString(vm_utils.ParamKeyPolicy, policy.ID)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
