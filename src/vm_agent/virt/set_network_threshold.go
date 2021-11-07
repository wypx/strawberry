package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyNetworkThresholdExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifyNetworkThresholdExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) error {
	guestID, err := request.GetString(VmUtils.ParamKeyGuest)
	if err != nil {
		return err
	}
	limitParameters, err := request.GetUIntArray(VmUtils.ParamKeyLimit)
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

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyNetworkThresholdResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var respChan = make(chan error, 1)
	executor.InstanceModule.ModifyNetworkThreshold(guestID, receiveSpeed, sendSpeed, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] modify network threshold fail: %s", id, err.Error())
		resp.SetError(err.Error())
	} else {
		log.Printf("[%08X] network threshold of guest '%s' changed to receive %d Kps, send %d Kps", id, guestID,
			receiveSpeed>>10, sendSpeed>>10)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
