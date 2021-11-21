package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryUnallocatedCellsExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryUnallocatedCellsExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {

	log.Printf("[%08X] query unallocated compute cells from %s.[%08X]", id, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)

	executor.ResourceModule.GetUnallocatedCells(respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryUnallocatedComputePoolCellResponse)
	resp.SetSuccess(true)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	modules.CellsToMessage(resp, result.ComputeCellInfoList)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
