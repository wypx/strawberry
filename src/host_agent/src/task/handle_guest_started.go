package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleGuestStartedExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *HandleGuestStartedExecutor) Execute(id vm_utils.SessionID, event vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := event.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	log.Printf("[%08X] recv guest '%s' started from %s.[%08X]", id, instanceID,
		event.GetSender(), event.GetFromSession())
	var status modules.InstanceStatus
	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(instanceID, respChan)
		result := <-respChan
		if result.Error != nil {
			errMsg := result.Error.Error()
			log.Printf("[%08X] fetch guest fail: %s", id, errMsg)
			return result.Error
		}
		status = result.Instance
	}
	status.Running = true
	{
		var respChan = make(chan error)
		executor.ResourceModule.UpdateInstanceStatus(status, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] warning: update started status fail: %s", id, err)
		}
		return nil
	}
}
