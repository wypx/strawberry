package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryZoneStatusExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryZoneStatusExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {

	//log.Printf("[%08X] query zone status from %s.[%08X]", id, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)

	executor.ResourceModule.QueryZoneStatus(respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryZoneStatusResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	if result.Error != nil {
		resp.SetSuccess(false)
		resp.SetError(result.Error.Error())
		log.Printf("[%08X] get zone status fail: %s", id, result.Error.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	//assemble
	var s = result.Zone
	resp.SetString(vm_utils.ParamKeyName, s.Name)
	resp.SetUIntArray(vm_utils.ParamKeyPool, []uint64{s.DisabledPools, s.EnabledPools})
	resp.SetUIntArray(vm_utils.ParamKeyCell, []uint64{s.OfflineCells, s.OnlineCells})
	resp.SetUIntArray(vm_utils.ParamKeyInstance, []uint64{s.StoppedInstances, s.RunningInstances, s.LostInstances, s.MigratingInstances})
	resp.SetFloat(vm_utils.ParamKeyUsage, s.CpuUsage)
	resp.SetUInt(vm_utils.ParamKeyCore, s.Cores)
	resp.SetUIntArray(vm_utils.ParamKeyMemory, []uint64{s.MemoryAvailable, s.Memory})
	resp.SetUIntArray(vm_utils.ParamKeyDisk, []uint64{s.DiskAvailable, s.Disk})
	resp.SetUIntArray(vm_utils.ParamKeySpeed, []uint64{s.ReadSpeed, s.WriteSpeed, s.ReceiveSpeed, s.SendSpeed})
	resp.SetString(vm_utils.ParamKeyStart, s.StartTime.Format(modules.TimeFormatLayout))

	return executor.Sender.SendMessage(resp, request.GetSender())
}
