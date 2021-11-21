package task

import (
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifyGuestPriorityExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifyGuestPriorityExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	guestID, err := request.GetString(vm_utils.ParamKeyGuest)
	if err != nil {
		return err
	}
	priorityValue, err := request.GetUInt(vm_utils.ParamKeyPriority)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request changing CPU priority of guest '%s' to %d from %s.[%08X]", id, guestID,
		priorityValue, request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyPriorityResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

	var ins modules.InstanceStatus

	{
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.GetInstanceStatus(guestID, respChan)
		result := <-respChan
		if result.Error != nil {
			log.Printf("[%08X] fetch instance fail: %s", id, result.Error.Error())
			resp.SetError(result.Error.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		ins = result.Instance
	}
	{
		//forward request
		forward, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyPriorityRequest)
		forward.SetFromSession(id)
		forward.SetString(vm_utils.ParamKeyGuest, guestID)
		forward.SetUInt(vm_utils.ParamKeyPriority, priorityValue)
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil {
			log.Printf("[%08X] forward modify CPU priority to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				var priority = modules.PriorityEnum(priorityValue)
				//update
				var respChan = make(chan error, 1)
				executor.ResourceModule.UpdateInstancePriority(guestID, priority, respChan)
				err = <-respChan
				if err != nil {
					log.Printf("[%08X] update CPU priority fail: %s", id, err.Error())
					resp.SetError(err.Error())
					return executor.Sender.SendMessage(resp, request.GetSender())
				}
				log.Printf("[%08X] modify CPU priority success", id)
			} else {
				log.Printf("[%08X] modify CPU priority fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait modify CPU priority response timeout", id)
			resp.SetError("request timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
