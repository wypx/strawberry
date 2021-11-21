package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetComputeCellStatusExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetComputeCellStatusExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err != nil {
		return err
	}

	//log.Printf("[%08X] get compute cell '%s' status from %s.[%08X]", id, cellName, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)

	executor.ResourceModule.GetComputeCellStatus(poolName, cellName, respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetComputePoolCellStatusResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	if result.Error != nil {
		err = result.Error
		resp.SetSuccess(false)
		resp.SetError(err.Error())
		log.Printf("[%08X] get compute cell status fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var s = result.ComputeCell

	resp.SetSuccess(true)
	//assemble
	resp.SetString(vm_utils.ParamKeyName, s.Name)
	resp.SetString(vm_utils.ParamKeyAddress, s.Address)
	resp.SetBoolean(vm_utils.ParamKeyEnable, s.Enabled)
	resp.SetBoolean(vm_utils.ParamKeyStatus, s.Alive)
	resp.SetUIntArray(vm_utils.ParamKeyInstance, []uint64{s.StoppedInstances, s.RunningInstances, s.LostInstances, s.MigratingInstances})
	resp.SetFloat(vm_utils.ParamKeyUsage, s.CpuUsage)
	resp.SetUInt(vm_utils.ParamKeyCore, s.Cores)
	resp.SetUIntArray(vm_utils.ParamKeyMemory, []uint64{s.MemoryAvailable, s.Memory})
	resp.SetUIntArray(vm_utils.ParamKeyDisk, []uint64{s.DiskAvailable, s.Disk})
	resp.SetUIntArray(vm_utils.ParamKeySpeed, []uint64{s.ReadSpeed, s.WriteSpeed, s.ReceiveSpeed, s.SendSpeed})

	return executor.Sender.SendMessage(resp, request.GetSender())
}
