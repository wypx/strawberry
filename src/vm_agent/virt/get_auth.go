package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type GetGuestPasswordExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *GetGuestPasswordExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	guestID, err := request.GetString(VmUtils.ParamKeyGuest)
	if err != nil {
		return err
	}

	//log.Printf("[%08X] request get password of '%s' from %s.[%08X]", id, guestID,
	//	request.GetSender(), request.GetFromSession())

	var respChan = make(chan VmAgentSvc.InstanceResult)
	executor.InstanceModule.GetGuestAuth(guestID, respChan)

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.GetAuthResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	result := <-respChan
	if result.Error != nil {
		resp.SetError(result.Error.Error())
		log.Printf("[%08X] get password fail: %s", id, result.Error.Error())
	} else {
		resp.SetSuccess(true)
		resp.SetString(VmUtils.ParamKeyUser, result.User)
		resp.SetString(VmUtils.ParamKeySecret, result.Password)

	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
