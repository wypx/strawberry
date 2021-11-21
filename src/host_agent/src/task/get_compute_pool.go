package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetComputePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetComputePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return
	}
	log.Printf("[%08X] get compute pool '%s' from %s.[%08X]", id, poolName, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)
	executor.ResourceModule.GetComputePool(poolName, respChan)
	result := <-respChan
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetComputePoolResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] get compute pool fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var poolInfo = result.ComputePoolConfig
	resp.SetString(vm_utils.ParamKeyName, poolInfo.Name)
	resp.SetBoolean(vm_utils.ParamKeyEnable, poolInfo.Enabled)
	resp.SetUInt(vm_utils.ParamKeyCell, uint(poolInfo.CellCount))
	resp.SetString(vm_utils.ParamKeyNetwork, poolInfo.Network)
	resp.SetString(vm_utils.ParamKeyStorage, poolInfo.Storage)
	resp.SetBoolean(vm_utils.ParamKeyOption, poolInfo.Failover)
	resp.SetSuccess(true)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
