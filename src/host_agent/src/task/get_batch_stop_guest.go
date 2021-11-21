package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetBatchStopGuestExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetBatchStopGuestExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var batchID string
	if batchID, err = request.GetString(vm_utils.ParamKeyID); err != nil {
		return err
	}
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetBatchStopGuestResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetBatchStopGuestStatus(batchID, respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] get batch stop status from %s.[%08X] fail: %s", id, request.GetSender(), request.GetFromSession(), err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	var guestStatus []uint64
	var guestID, guestName, stopError []string

	for _, status := range result.BatchStop {
		guestStatus = append(guestStatus, uint64(status.Status))
		guestID = append(guestID, status.ID)
		guestName = append(guestName, status.Name)
		stopError = append(stopError, status.Error)
	}
	resp.SetSuccess(true)
	resp.SetStringArray(vm_utils.ParamKeyName, guestName)
	resp.SetStringArray(vm_utils.ParamKeyGuest, guestID)
	resp.SetStringArray(vm_utils.ParamKeyError, stopError)
	resp.SetUIntArray(vm_utils.ParamKeyStatus, guestStatus)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
