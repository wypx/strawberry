package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetSecurityPolicyGroupExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetSecurityPolicyGroupExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var policyID string
	if policyID, err = request.GetString(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy group ID fail: %s", err.Error())
		return
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetPolicyGroupResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetSecurityPolicyGroup(policyID, respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] get security policy group '%s' fail: %s",
			id, policyID, err.Error())
		resp.SetError(err.Error())
	} else {
		var policy = result.PolicyGroup
		resp.SetString(vm_utils.ParamKeyPolicy, policy.ID)
		resp.SetString(vm_utils.ParamKeyName, policy.Name)
		resp.SetString(vm_utils.ParamKeyDescription, policy.Description)
		resp.SetString(vm_utils.ParamKeyUser, policy.User)
		resp.SetString(vm_utils.ParamKeyGroup, policy.Group)
		resp.SetBoolean(vm_utils.ParamKeyAction, policy.Accept)
		resp.SetBoolean(vm_utils.ParamKeyEnable, policy.Enabled)
		resp.SetBoolean(vm_utils.ParamKeyLimit, policy.Global)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
