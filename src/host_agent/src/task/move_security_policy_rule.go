package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type MoveSecurityPolicyRuleExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *MoveSecurityPolicyRuleExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var policyID string
	if policyID, err = request.GetString(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy group ID fail: %s", err.Error())
		return
	}
	var index int
	if index, err = request.GetInt(vm_utils.ParamKeyIndex); err != nil {
		err = fmt.Errorf("get index fail: %s", err.Error())
		return
	}
	var moveUp bool
	if moveUp, err = request.GetBoolean(vm_utils.ParamKeyFlag); err != nil {
		err = fmt.Errorf("get move flag fail: %s", err.Error())
		return
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ChangePolicyRuleOrderResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.ResourceModule.MoveSecurityPolicyRule(policyID, index, moveUp, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] move %dth rule of security policy '%s' fail: %s",
			id, index, policyID, err.Error())
		resp.SetError(err.Error())
	} else {
		if moveUp {
			log.Printf("[%08X] %dth rule of security policy '%s' moved up",
				id, index, policyID)
		} else {
			log.Printf("[%08X] %dth rule of security policy '%s' moved down",
				id, index, policyID)
		}
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
