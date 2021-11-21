package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type RemoveGuestSecurityRuleExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *RemoveGuestSecurityRuleExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var instanceID string
	if instanceID, err = request.GetString(vm_utils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance ID fail: %s", err.Error())
		return
	}
	var index int
	if index, err = request.GetInt(vm_utils.ParamKeyIndex); err != nil {
		err = fmt.Errorf("get index fail: %s", err.Error())
		return
	}
	var instance modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.RemoveGuestRuleResponse)
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
			log.Printf("[%08X] get instance '%s' for remove security rule fail: %s",
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
			log.Printf("[%08X] forward remove security rule to cell '%s' fail: %s", id, instance.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				log.Printf("[%08X] %dth security rule of instance '%s' removed",
					id, index, instance.Name)
			} else {
				log.Printf("[%08X] cell remove security rule fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			cellResp.SetTransactionID(request.GetTransactionID())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait remove security rule response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
