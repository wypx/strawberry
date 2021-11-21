package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetComputePoolStatusExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetComputePoolStatusExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}

	//log.Printf("[%08X] get compute pool '%s' status from %s.[%08X]", id, poolName, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)

	executor.ResourceModule.GetComputePoolStatus(poolName, respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetComputePoolStatusResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	if result.Error != nil {
		err = result.Error
		resp.SetSuccess(false)
		resp.SetError(err.Error())
		log.Printf("[%08X] get compute pool status fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var s = result.ComputePool

	resp.SetSuccess(true)
	//assemble
	resp.SetString(vm_utils.ParamKeyName, s.Name)
	resp.SetBoolean(vm_utils.ParamKeyEnable, s.Enabled)
	resp.SetUIntArray(vm_utils.ParamKeyCell, []uint64{s.OfflineCells, s.OnlineCells})
	resp.SetUIntArray(vm_utils.ParamKeyInstance, []uint64{s.StoppedInstances, s.RunningInstances, s.LostInstances, s.MigratingInstances})
	resp.SetFloat(vm_utils.ParamKeyUsage, s.CpuUsage)
	resp.SetUInt(vm_utils.ParamKeyCore, s.Cores)
	resp.SetUIntArray(vm_utils.ParamKeyMemory, []uint64{s.MemoryAvailable, s.Memory})
	resp.SetUIntArray(vm_utils.ParamKeyDisk, []uint64{s.DiskAvailable, s.Disk})
	resp.SetUIntArray(vm_utils.ParamKeySpeed, []uint64{s.ReadSpeed, s.WriteSpeed, s.ReceiveSpeed, s.SendSpeed})

	return executor.Sender.SendMessage(resp, request.GetSender())
}
