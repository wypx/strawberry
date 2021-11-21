package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleGuestDeletedExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *HandleGuestDeletedExecutor) Execute(id vm_utils.SessionID, event vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := event.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	log.Printf("[%08X] recv guest '%s' deleted from %s.[%08X]", id, instanceID,
		event.GetSender(), event.GetFromSession())
	var respChan = make(chan error)
	executor.ResourceModule.DeallocateInstance(instanceID, nil, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] deallocate guest fail: %s", id, err.Error())
	}
	return nil
}
