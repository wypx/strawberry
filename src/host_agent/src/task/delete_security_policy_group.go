package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type DeleteSecurityPolicyGroupExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *DeleteSecurityPolicyGroupExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var policyID string
	if policyID, err = request.GetString(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy group ID fail: %s", err.Error())
		return
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DeletePolicyGroupResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.ResourceModule.DeleteSecurityPolicyGroup(policyID, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] delete security policy group '%s' fail: %s",
			id, policyID, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] security policy group '%s' deleted",
			id, policyID)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
