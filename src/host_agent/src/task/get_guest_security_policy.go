package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetGuestSecurityPolicyExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetGuestSecurityPolicyExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var instanceID string
	if instanceID, err = request.GetString(vm_utils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance ID fail: %s", err.Error())
		return
	}
	var instance modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetGuestRuleResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	{
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.GetInstanceStatus(instanceID, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] get instance '%s' for security policy fail: %s",
				id, instanceID, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		instance = result.Instance
	}
	{
		//forward request
		var forward = vm_utils.CloneJsonMessage(request)
		forward.SetFromSession(id)
		if err = executor.Sender.SendMessage(forward, instance.Cell); err != nil {
			log.Printf("[%08X] forward get secuirty policy to cell '%s' fail: %s", id, instance.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if !cellResp.IsSuccess() {
				log.Printf("[%08X] cell get secuirty policy fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			cellResp.SetTransactionID(request.GetTransactionID())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait get secuirty policy timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
