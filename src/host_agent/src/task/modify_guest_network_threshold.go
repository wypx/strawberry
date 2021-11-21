package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifyGuestNetworkThresholdExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifyGuestNetworkThresholdExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var guestID string
	guestID, err = request.GetString(vm_utils.ParamKeyGuest)
	if err != nil {
		return err
	}
	limitParameters, err := request.GetUIntArray(vm_utils.ParamKeyLimit)
	if err != nil {
		return err
	}
	const (
		ReceiveOffset = iota
		SendOffset
		ValidLimitParametersCount = 2
	)

	if ValidLimitParametersCount != len(limitParameters) {
		var err = fmt.Errorf("invalid QoS parameters count %d", len(limitParameters))
		return err
	}
	var receiveSpeed = limitParameters[ReceiveOffset]
	var sendSpeed = limitParameters[SendOffset]

	log.Printf("[%08X] request modifying network threshold of guest '%s' from %s.[%08X]", id, guestID,
		request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyNetworkThresholdResponse)
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
		forward, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyNetworkThresholdRequest)
		forward.SetFromSession(id)
		forward.SetString(vm_utils.ParamKeyGuest, guestID)
		forward.SetUIntArray(vm_utils.ParamKeyLimit, []uint64{receiveSpeed, sendSpeed})
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil {
			log.Printf("[%08X] forward modify network threshold to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				//update
				var respChan = make(chan error, 1)
				executor.ResourceModule.UpdateInstanceNetworkThreshold(guestID, receiveSpeed, sendSpeed, respChan)
				err = <-respChan
				if err != nil {
					log.Printf("[%08X] update network threshold fail: %s", id, err.Error())
					resp.SetError(err.Error())
					return executor.Sender.SendMessage(resp, request.GetSender())
				}
				log.Printf("[%08X] modify network threshold success", id)
			} else {
				log.Printf("[%08X] modify network threshold fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait modify network threshold response timeout", id)
			resp.SetError("request timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
