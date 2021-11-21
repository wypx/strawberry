package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleAddressChangedExecutor struct {
	ResourceModule modules.ResourceModule
}

func (executor *HandleAddressChangedExecutor) Execute(id vm_utils.SessionID, event vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := event.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	address, err := event.GetString(vm_utils.ParamKeyAddress)
	if err != nil {
		return err
	}

	log.Printf("[%08X] address of guest '%s' changed to %s, notify from %s.[%08X]", id, instanceID,
		address, event.GetSender(), event.GetFromSession())
	var respChan = make(chan error)
	executor.ResourceModule.UpdateInstanceAddress(instanceID, address, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] update address fail: %s", id, err.Error())
	}
	return nil
}
