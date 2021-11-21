package task

import (
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryComputePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryComputePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	//log.Printf("[%08X] query compute pool from %s.[%08X]", id, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)
	executor.ResourceModule.GetAllComputePool(respChan)
	result := <-respChan
	var nameArray, networkArray, storageArray []string
	var cellArray, statusArray, failoverArray []uint64
	for _, info := range result.ComputePoolInfoList {
		if info.Enabled {
			statusArray = append(statusArray, 1)
		} else {
			statusArray = append(statusArray, 0)
		}
		if info.Failover {
			failoverArray = append(failoverArray, 1)
		} else {
			failoverArray = append(failoverArray, 0)
		}
		nameArray = append(nameArray, info.Name)
		cellArray = append(cellArray, info.CellCount)
		networkArray = append(networkArray, info.Network)
		storageArray = append(storageArray, info.Storage)
	}
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryComputePoolResponse)
	resp.SetSuccess(true)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetStringArray(vm_utils.ParamKeyName, nameArray)
	resp.SetStringArray(vm_utils.ParamKeyNetwork, networkArray)
	resp.SetStringArray(vm_utils.ParamKeyStorage, storageArray)
	resp.SetUIntArray(vm_utils.ParamKeyStatus, statusArray)
	resp.SetUIntArray(vm_utils.ParamKeyCell, cellArray)
	resp.SetUIntArray(vm_utils.ParamKeyOption, failoverArray)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
