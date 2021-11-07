package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type HandleComputePoolReadyExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
	NetworkModule  VmAgentSvc.NetworkModule
}

func (executor *HandleComputePoolReadyExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(VmUtils.ParamKeyPool)
	if err != nil {
		return err
	}
	storageName, err := request.GetString(VmUtils.ParamKeyStorage)
	if err != nil {
		return
	}
	networkName, err := request.GetString(VmUtils.ParamKeyNetwork)
	if err != nil {
		return
	}
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ComputeCellReadyEvent)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	if "" == storageName {
		log.Printf("[%08X] recv compute pool '%s' ready from %s", id, poolName, request.GetSender())
		//try detach
		var respChan = make(chan error, 1)
		executor.StorageModule.DetachStorage(respChan)
		err = <-respChan
		if err != nil {
			resp.SetError(err.Error())
			log.Printf("[%08X] detach storage fail: %s", id, err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	} else {
		var protocol, host, target string
		if protocol, err = request.GetString(VmUtils.ParamKeyType); err != nil {
			return
		}
		if host, err = request.GetString(VmUtils.ParamKeyHost); err != nil {
			return
		}
		if target, err = request.GetString(VmUtils.ParamKeyTarget); err != nil {
			return
		}
		log.Printf("[%08X] recv compute pool '%s' ready using storage '%s' from %s", id, poolName, storageName, request.GetSender())
		var storageURL string
		{
			var respChan = make(chan VmAgentSvc.StorageResult, 1)
			executor.StorageModule.UsingStorage(storageName, protocol, host, target, respChan)
			var result = <-respChan
			if result.Error != nil {
				err = result.Error
				resp.SetError(err.Error())
				log.Printf("[%08X] using storage fail: %s", id, err.Error())
				return executor.Sender.SendMessage(resp, request.GetSender())
			}
			//storage ready
			storageURL = result.Path
		}
		{
			var respChan = make(chan error, 1)
			executor.InstanceModule.UsingStorage(storageName, storageURL, respChan)
			err = <-respChan
			if err != nil {
				resp.SetError(err.Error())
				log.Printf("[%08X] update storage URL of instance to '%s' fail: %s", id, storageURL, err.Error())
				return executor.Sender.SendMessage(resp, request.GetSender())
			}
		}

	}
	var allocationMode string
	if "" != networkName {
		var gateway string
		var dns []string
		if gateway, err = request.GetString(VmUtils.ParamKeyGateway); err != nil {
			return
		}

		if dns, err = request.GetStringArray(VmUtils.ParamKeyServer); err != nil {
			return
		}
		if allocationMode, err = request.GetString(VmUtils.ParamKeyMode); err != nil {
			err = fmt.Errorf("get allocation mode fail: %s", err.Error())
			return
		}
		switch allocationMode {
		case VmAgentSvc.AddressAllocationNone:
		case VmAgentSvc.AddressAllocationDHCP:
		case VmAgentSvc.AddressAllocationCloudInit:
			break
		default:
			err = fmt.Errorf("invalid allocation mode :%s", allocationMode)
			return
		}
		var respChan = make(chan error, 1)
		executor.NetworkModule.UpdateAddressAllocation(gateway, dns, allocationMode, respChan)
		err = <-respChan
		if err != nil {
			resp.SetError(err.Error())
			log.Printf("[%08X] update address allocation fail: %s", id, err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	if VmAgentSvc.AddressAllocationNone != allocationMode {
		executor.InstanceModule.SyncAddressAllocation(allocationMode)
	}
	var respChan = make(chan []VmAgentSvc.GuestConfig)
	executor.InstanceModule.GetAllInstance(respChan)
	allConfig := <-respChan
	var count = uint(len(allConfig))

	resp.SetSuccess(true)
	resp.SetUInt(VmUtils.ParamKeyCount, count)
	if 0 == count {
		log.Printf("[%08X] no instance configured", id)
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	var names, ids, users, groups, secrets, addresses, systems, createTime, internal, external, hardware []string
	var cores, options, enables, progress, status, monitors, memories, disks, diskCounts, cpuPriorities, ioLimits []uint64
	for _, config := range allConfig {
		names = append(names, config.Name)
		ids = append(ids, config.ID)
		users = append(users, config.User)
		groups = append(groups, config.Group)
		cores = append(cores, uint64(config.Cores))
		if config.AutoStart {
			options = append(options, 1)
		} else {
			options = append(options, 0)
		}
		if config.Created {
			enables = append(enables, 1)
			progress = append(progress, 0)
		} else {
			enables = append(enables, 0)
			progress = append(progress, uint64(config.Progress))
		}
		if config.Running {
			status = append(status, VmAgentSvc.InstanceStatusRunning)
		} else {
			status = append(status, VmAgentSvc.InstanceStatusStopped)
		}
		monitors = append(monitors, uint64(config.MonitorPort))
		secrets = append(secrets, config.MonitorSecret)
		memories = append(memories, uint64(config.Memory))
		var diskCount = len(config.Disks)
		diskCounts = append(diskCounts, uint64(diskCount))
		for _, diskSize := range config.Disks {
			disks = append(disks, diskSize)
		}
		addresses = append(addresses, config.NetworkAddress)
		var operatingSystem string
		if nil != config.Template {
			operatingSystem = config.Template.OperatingSystem
		} else {
			operatingSystem = config.System
		}
		systems = append(systems, operatingSystem)
		createTime = append(createTime, config.CreateTime)
		internal = append(internal, config.InternalAddress)
		external = append(external, config.ExternalAddress)
		hardware = append(hardware, config.HardwareAddress)
		cpuPriorities = append(cpuPriorities, uint64(config.CPUPriority))
		ioLimits = append(ioLimits, []uint64{config.ReadSpeed, config.WriteSpeed,
			config.ReadIOPS, config.WriteIOPS, config.ReceiveSpeed, config.SendSpeed}...)
	}
	resp.SetStringArray(VmUtils.ParamKeyName, names)
	resp.SetStringArray(VmUtils.ParamKeyInstance, ids)
	resp.SetStringArray(VmUtils.ParamKeyUser, users)
	resp.SetStringArray(VmUtils.ParamKeyGroup, groups)
	resp.SetStringArray(VmUtils.ParamKeySecret, secrets)
	resp.SetStringArray(VmUtils.ParamKeyAddress, addresses)
	resp.SetStringArray(VmUtils.ParamKeySystem, systems)
	resp.SetStringArray(VmUtils.ParamKeyCreate, createTime)
	resp.SetStringArray(VmUtils.ParamKeyInternal, internal)
	resp.SetStringArray(VmUtils.ParamKeyExternal, external)
	resp.SetStringArray(VmUtils.ParamKeyHardware, hardware)
	resp.SetUIntArray(VmUtils.ParamKeyCore, cores)
	resp.SetUIntArray(VmUtils.ParamKeyOption, options)
	resp.SetUIntArray(VmUtils.ParamKeyEnable, enables)
	resp.SetUIntArray(VmUtils.ParamKeyProgress, progress)
	resp.SetUIntArray(VmUtils.ParamKeyStatus, status)
	resp.SetUIntArray(VmUtils.ParamKeyMonitor, monitors)
	resp.SetUIntArray(VmUtils.ParamKeyMemory, memories)
	resp.SetUIntArray(VmUtils.ParamKeyCount, diskCounts)
	resp.SetUIntArray(VmUtils.ParamKeyDisk, disks)
	resp.SetUIntArray(VmUtils.ParamKeyPriority, cpuPriorities)
	resp.SetUIntArray(VmUtils.ParamKeyLimit, ioLimits)
	log.Printf("[%08X] %d instance config(s) reported", id, count)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
