package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type RemoveComputePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *RemoveComputePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	pool, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request remove cell '%s' from pool '%s' from %s.[%08X]", id, cellName, pool,
		request.GetSender(), request.GetFromSession())
	var respChan = make(chan error)
	executor.ResourceModule.RemoveCell(pool, cellName, respChan)

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.RemoveComputePoolCellResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] remove compute cell fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: send cell removed response to '%s' fail: %s", id, request.GetSender(), err.Error())
	}
	event, _ := vm_utils.CreateJsonMessage(vm_utils.ComputeCellRemovedEvent)
	if err = executor.Sender.SendMessage(event, cellName); err != nil {
		log.Printf("[%08X] warning: notify cell removed to '%s' fail: %s", id, cellName, err.Error())
	}
	return nil
}
