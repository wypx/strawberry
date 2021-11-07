package virt

import (
	"errors"
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ResetMonitorSecretExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ResetMonitorSecretExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var guestID string
	if guestID, err = request.GetString(VmUtils.ParamKeyGuest); err != nil {
		err = fmt.Errorf("get guest id fail: %s", err.Error())
		return err
	}
	var respChan = make(chan VmAgentSvc.InstanceResult)
	executor.InstanceModule.ResetMonitorPassword(guestID, respChan)

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ResetSecretResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var password string
	result := <-respChan
	if result.Error != nil {
		err = result.Error

	} else {
		password = result.Password
		if "" == password {
			err = errors.New("new password is empty")
		}
	}
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] reset monitor secret fail: %s", id, err.Error())
	} else {
		resp.SetSuccess(true)
		resp.SetString(VmUtils.ParamKeySecret, password)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
