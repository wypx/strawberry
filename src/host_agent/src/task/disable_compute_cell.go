package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type DisableComputeCellExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *DisableComputeCellExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err != nil {
		return err
	}
	var respChan = make(chan error, 1)
	executor.ResourceModule.DisableCell(poolName, cellName, false, respChan)

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DisableComputePoolCellResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] disable compute cell fail: %s", id, err.Error())
	} else {
		resp.SetSuccess(true)
		log.Printf("[%08X] cell '%s' disabled in pool %s", id, cellName, poolName)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
