package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleGuestCreatedExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *HandleGuestCreatedExecutor) Execute(id vm_utils.SessionID, event vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var instanceID, monitorSecret, ethernet string
	var monitorPort uint
	if instanceID, err = event.GetString(vm_utils.ParamKeyInstance); err != nil {
		return
	}

	if monitorPort, err = event.GetUInt(vm_utils.ParamKeyMonitor); err != nil {
		return
	}

	if monitorSecret, err = event.GetString(vm_utils.ParamKeySecret); err != nil {
		return
	}
	if ethernet, err = event.GetString(vm_utils.ParamKeyHardware); err != nil {
		return
	}
	log.Printf("[%08X] recv guest '%s' created from %s.[%08X], monitor port %d", id, instanceID,
		event.GetSender(), event.GetFromSession(), monitorPort)
	var respChan = make(chan error)
	executor.ResourceModule.ConfirmInstance(instanceID, monitorPort, monitorSecret, ethernet, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] confirm instance fail: %s", id, err.Error())
	}
	return nil
}
