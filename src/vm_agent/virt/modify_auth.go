package virt

import (
	"log"
	"math/rand"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifyGuestPasswordExecutor struct {
	Sender          VmUtils.MessageSender
	InstanceModule  VmAgentSvc.InstanceModule
	RandomGenerator *rand.Rand
}

func (executor *ModifyGuestPasswordExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	const (
		PasswordLength = 10
	)
	var guestID, user, password string

	if guestID, err = request.GetString(VmUtils.ParamKeyGuest); err != nil {
		return err
	}
	if user, err = request.GetString(VmUtils.ParamKeyUser); err != nil {
		return err
	}
	if password, err = request.GetString(VmUtils.ParamKeySecret); err != nil {
		return err
	}

	if "" == password {
		password = executor.generatePassword(PasswordLength)
		log.Printf("[%08X] new password '%s' generated for modify auth", id, password)
	}

	var respChan = make(chan VmAgentSvc.InstanceResult)
	executor.InstanceModule.ModifyGuestAuth(guestID, password, user, respChan)

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyAuthResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	result := <-respChan
	if result.Error != nil {
		resp.SetError(result.Error.Error())
		log.Printf("[%08X] modify password fail: %s", id, result.Error.Error())
	} else {
		resp.SetSuccess(true)
		resp.SetString(VmUtils.ParamKeyUser, result.User)
		resp.SetString(VmUtils.ParamKeySecret, result.Password)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}

func (executor *ModifyGuestPasswordExecutor) generatePassword(length int) string {
	const (
		Letters = "~!@#$%^&*()_[]-=+0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	)
	var result = make([]byte, length)
	var n = len(Letters)
	for i := 0; i < length; i++ {
		result[i] = Letters[executor.RandomGenerator.Intn(n)]
	}
	return string(result)
}
