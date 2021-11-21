package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryComputeCellStatusExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryComputeCellStatusExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}

	//log.Printf("[%08X] query cell status in pool '%s' from %s.[%08X]", id, poolName, request.GetSender(), request.GetFromSession())

	var respChan = make(chan modules.ResourceResult)
	executor.ResourceModule.QueryComputeCellStatus(poolName, respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryComputePoolCellStatusResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetSuccess(false)
		resp.SetError(err.Error())
		log.Printf("[%08X] query compute cell status fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	resp.SetSuccess(true)
	var name, address []string
	var enabled, alive, instance, usage, cores, memory, disk, speed []uint64
	for _, s := range result.ComputeCellList {
		name = append(name, s.Name)
		address = append(address, s.Address)
		if s.Enabled {
			enabled = append(enabled, 1)
		} else {
			enabled = append(enabled, 0)
		}
		if s.Alive {
			alive = append(alive, 1)
		} else {
			alive = append(alive, 0)
		}
		instance = append(instance, s.StoppedInstances)
		instance = append(instance, s.RunningInstances)
		instance = append(instance, s.LostInstances)
		instance = append(instance, s.MigratingInstances)
		usage = append(usage, uint64(s.CpuUsage)) //todo: tripped decimal
		cores = append(cores, uint64(s.Cores))
		memory = append(memory, s.MemoryAvailable)
		memory = append(memory, s.Memory)
		disk = append(disk, s.DiskAvailable)
		disk = append(disk, s.Disk)
		speed = append(speed, s.ReadSpeed)
		speed = append(speed, s.WriteSpeed)
		speed = append(speed, s.ReceiveSpeed)
		speed = append(speed, s.SendSpeed)
	}

	//assemble
	resp.SetStringArray(vm_utils.ParamKeyName, name)
	resp.SetStringArray(vm_utils.ParamKeyAddress, address)
	resp.SetUIntArray(vm_utils.ParamKeyEnable, enabled)
	resp.SetUIntArray(vm_utils.ParamKeyStatus, alive)
	resp.SetUIntArray(vm_utils.ParamKeyInstance, instance)
	resp.SetUIntArray(vm_utils.ParamKeyUsage, usage)
	resp.SetUIntArray(vm_utils.ParamKeyCore, cores)
	resp.SetUIntArray(vm_utils.ParamKeyMemory, memory)
	resp.SetUIntArray(vm_utils.ParamKeyDisk, disk)
	resp.SetUIntArray(vm_utils.ParamKeySpeed, speed)
	//log.Printf("[%08X] %d compute cell status available", id, len(name))
	return executor.Sender.SendMessage(resp, request.GetSender())
}
