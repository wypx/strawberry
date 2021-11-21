package task

import (
	"errors"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleGuestUpdatedExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *HandleGuestUpdatedExecutor) Execute(id vm_utils.SessionID, event vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := event.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	if !event.IsSuccess() {
		log.Printf("[%08X] guest '%s' create fail: %s", id, instanceID, event.GetError())
		err = errors.New(event.GetError())
		var respChan = make(chan error)
		executor.ResourceModule.DeallocateInstance(instanceID, err, respChan)
		<-respChan
		return nil
	}

	progress, err := event.GetUInt(vm_utils.ParamKeyProgress)
	if err != nil {
		return err
	}

	log.Printf("[%08X] update guest '%s' progress to %d%% from %s.[%08X]", id, instanceID,
		progress, event.GetSender(), event.GetFromSession())

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
		if result.Instance.Created {
			log.Printf("[%08X] warning: guest already created", id)
			return nil
		}
		status = result.Instance
	}
	status.Progress = progress
	{
		var respChan = make(chan error)
		executor.ResourceModule.UpdateInstanceStatus(status, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] warning: update progress fail: %s", id, err)
		}
		return nil
	}
}
