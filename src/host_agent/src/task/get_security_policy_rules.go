package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetSecurityPolicyRulesExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetSecurityPolicyRulesExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var policyID string
	if policyID, err = request.GetString(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy group ID fail: %s", err.Error())
		return
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryPolicyRuleResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetSecurityPolicyRules(policyID, respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] get all rules of security policy '%s' fail: %s",
			id, policyID, err.Error())
		resp.SetError(err.Error())
	} else {
		var actions, targetPorts []uint64
		var protocols, sourceAddresses, targetAddresses []string
		for _, rule := range result.PolicyRuleList {
			if rule.Accept {
				actions = append(actions, modules.PolicyRuleActionAccept)
			} else {
				actions = append(actions, modules.PolicyRuleActionReject)
			}
			targetPorts = append(targetPorts, uint64(rule.TargetPort))
			protocols = append(protocols, string(rule.Protocol))
			targetAddresses = append(targetAddresses, rule.TargetAddress)
			sourceAddresses = append(sourceAddresses, rule.SourceAddress)
		}
		resp.SetUIntArray(vm_utils.ParamKeyAction, actions)
		resp.SetUIntArray(vm_utils.ParamKeyPort, targetPorts)
		resp.SetStringArray(vm_utils.ParamKeyProtocol, protocols)
		resp.SetStringArray(vm_utils.ParamKeyFrom, sourceAddresses)
		resp.SetStringArray(vm_utils.ParamKeyTo, targetAddresses)
		log.Printf("[%08X] %d rules of security policy '%s' available",
			id, len(actions), policyID)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
