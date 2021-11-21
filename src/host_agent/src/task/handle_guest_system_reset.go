package task

import (
	"errors"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleGuestSystemResetExecutor struct {
	ResourceModule modules.ResourceModule
}

func (executor *HandleGuestSystemResetExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var guestID string
	if guestID, err = request.GetString(vm_utils.ParamKeyGuest); err != nil {
		return
	}
	var respChan = make(chan error, 1)
	if !request.IsSuccess() {
		err = errors.New(request.GetError())
	}
	executor.ResourceModule.FinishResetSystem(guestID, err, respChan)
	err = <-respChan
	if err != nil {
		log.Printf("[%08X] recv guest reset finish, but update fail: %s", id, err.Error())
	} else {
		log.Printf("[%08X] reset system of guest '%s' finished", id, guestID)
	}
	return nil
}
