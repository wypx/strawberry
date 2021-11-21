package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetGuestConfigExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetGuestConfigExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	var err error
	instanceID, err := request.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	//log.Printf("[%08X] request get guest '%s' config from %s.[%08X]", id, instanceID,
	//	request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryGuestResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

	var config modules.InstanceStatus
	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(instanceID, respChan)
		result := <-respChan
		if result.Error != nil {
			errMsg := result.Error.Error()
			log.Printf("[%08X] get config fail: %s", id, errMsg)
			resp.SetError(errMsg)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		config = result.Instance
	}
	config.Marshal(resp)
	resp.SetSuccess(true)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
