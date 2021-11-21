package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type EnableComputeCellExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *EnableComputeCellExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err != nil {
		return err
	}
	//log.Printf("[%08X] request enable cell '%s' in pool '%s' from %s.[%08X]", id, cellName, poolName,
	//	request.GetSender(), request.GetFromSession())
	var respChan = make(chan error, 1)
	executor.ResourceModule.EnableCell(poolName, cellName, respChan)

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.EnableComputePoolCellResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] enable compute cell fail: %s", id, err.Error())
	} else {
		resp.SetSuccess(true)
		log.Printf("[%08X] cell '%s' enabled in pool %s", id, cellName, poolName)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
