package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type DeleteGuestExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *DeleteGuestExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := request.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request delete guest '%s' from %s.[%08X]", id, instanceID,
		request.GetSender(), request.GetFromSession())
	var ins modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteGuestResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(instanceID, respChan)
		result := <-respChan
		if result.Error != nil {
			log.Printf("[%08X] fetch instance fail: %s", id, result.Error.Error())
			resp.SetError(result.Error.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		ins = result.Instance
		if ins.Running {
			err = fmt.Errorf("instance '%s' is still running", instanceID)
			log.Printf("[%08X] instance '%s' is still running", id, instanceID)
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		//request delete
		forward, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteGuestRequest)
		forward.SetFromSession(id)
		forward.SetString(vm_utils.ParamKeyInstance, instanceID)
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil {
			log.Printf("[%08X] forward delete to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				log.Printf("[%08X] cell delete guest success", id)
			} else {
				log.Printf("[%08X] cell delete guest fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			cellResp.SetTransactionID(request.GetTransactionID())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait delete response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
