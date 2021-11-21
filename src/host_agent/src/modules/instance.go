package modules

import (
	"fmt"
	"vm_manager/vm_utils"
)

type InstanceResource struct {
	Cores  uint
	Memory uint
	Disks  []uint64
}

type PriorityEnum uint

const (
	PriorityHigh = iota
	PriorityMedium
	PriorityLow
)

type InstanceNetworkInfo struct {
	InstanceAddress string
	MonitorAddress  string
	AssignedAddress string
	MonitorPort     uint
	MappedPorts     map[int]int
}

type InstanceStatus struct {
	InstanceResource
	Name            string
	ID              string
	Pool            string
	Cell            string
	Host            string //hosting cell ip
	User            string
	Group           string
	AutoStart       bool
	System          string
	Created         bool
	Progress        uint //limit to 100
	Running         bool
	Lost            bool
	Migrating       bool
	InternalNetwork InstanceNetworkInfo
	ExternalNetwork InstanceNetworkInfo
	MediaAttached   bool
	MediaSource     string
	MediaName       string
	MonitorProtocol string
	MonitorSecret   string
	CreateTime      string
	HardwareAddress string
	CPUPriority     PriorityEnum
	WriteSpeed      uint64
	WriteIOPS       uint64
	ReadSpeed       uint64
	ReadIOPS        uint64
	ReceiveSpeed    uint64
	SendSpeed       uint64
}

const (
	InstanceStatusStopped = iota
	InstanceStatusRunning
)

const (
	//bit 0~1 for running/stopped
	InstanceStatusLostBit    = 2
	InstanceStatusMigrateBit = 3
)

const (
	InstanceMediaOptionNone uint = iota
	InstanceMediaOptionImage
	InstanceMediaOptionNetwork
)

const (
	NetworkModePrivate = iota
	NetworkModePlain
	NetworkModeMono
	NetworkModeShare
	NetworkModeVPC
)

const (
	StorageModeLocal = iota
)

func MarshalInstanceStatusListToMessage(list []InstanceStatus, msg vm_utils.Message) error {
	var count = uint(len(list))
	msg.SetUInt(vm_utils.ParamKeyCount, count)
	var names, ids, pools, cells, hosts, users, monitors, addresses, groups, secrets, systems,
		createTime, internal, external, hardware []string
	var cores, options, enables, progress, status, memories, disks, diskCounts, mediaAttached, cpuPriorities, ioLimits []uint64
	for _, ins := range list {
		names = append(names, ins.Name)
		ids = append(ids, ins.ID)
		pools = append(pools, ins.Pool)
		cells = append(cells, ins.Cell)
		hosts = append(hosts, ins.Host)
		users = append(users, ins.User)
		groups = append(groups, ins.Group)
		cores = append(cores, uint64(ins.Cores))
		if ins.AutoStart {
			options = append(options, 1)
		} else {
			options = append(options, 0)
		}
		if ins.MediaAttached {
			mediaAttached = append(mediaAttached, 1)
		} else {
			mediaAttached = append(mediaAttached, 0)
		}
		if ins.Created {
			enables = append(enables, 1)
			progress = append(progress, 0)
		} else {
			enables = append(enables, 0)
			progress = append(progress, uint64(ins.Progress))
		}
		var insStatus uint64
		if ins.Running {
			insStatus = InstanceStatusRunning
		} else {
			insStatus = InstanceStatusStopped
		}
		if ins.Lost {
			insStatus |= 1 << InstanceStatusLostBit
		}
		status = append(status, insStatus)

		secrets = append(secrets, ins.MonitorSecret)
		var internalMonitor = fmt.Sprintf("%s:%d", ins.InternalNetwork.MonitorAddress, ins.InternalNetwork.MonitorPort)
		var externalMonitor = fmt.Sprintf("%s:%d", ins.ExternalNetwork.MonitorAddress, ins.ExternalNetwork.MonitorPort)
		monitors = append(monitors, internalMonitor)
		monitors = append(monitors, externalMonitor)
		addresses = append(addresses, ins.InternalNetwork.InstanceAddress)
		addresses = append(addresses, ins.ExternalNetwork.InstanceAddress)

		internal = append(internal, ins.InternalNetwork.AssignedAddress)
		external = append(external, ins.ExternalNetwork.AssignedAddress)

		systems = append(systems, ins.System)
		createTime = append(createTime, ins.CreateTime)
		hardware = append(hardware, ins.HardwareAddress)
		memories = append(memories, uint64(ins.Memory))
		var diskCount = len(ins.Disks)
		diskCounts = append(diskCounts, uint64(diskCount))
		for _, diskSize := range ins.Disks {
			disks = append(disks, diskSize)
		}
		//QoS
		cpuPriorities = append(cpuPriorities, uint64(ins.CPUPriority))
		ioLimits = append(ioLimits, []uint64{ins.ReadSpeed, ins.WriteSpeed,
			ins.ReadIOPS, ins.WriteIOPS, ins.ReceiveSpeed, ins.SendSpeed}...)
	}

	msg.SetStringArray(vm_utils.ParamKeyName, names)
	msg.SetStringArray(vm_utils.ParamKeyInstance, ids)
	msg.SetStringArray(vm_utils.ParamKeyPool, pools)
	msg.SetStringArray(vm_utils.ParamKeyCell, cells)
	msg.SetStringArray(vm_utils.ParamKeyHost, hosts)
	msg.SetStringArray(vm_utils.ParamKeyUser, users)

	msg.SetStringArray(vm_utils.ParamKeyMonitor, monitors)
	msg.SetStringArray(vm_utils.ParamKeySecret, secrets)
	msg.SetStringArray(vm_utils.ParamKeyAddress, addresses)
	msg.SetStringArray(vm_utils.ParamKeySystem, systems)
	msg.SetStringArray(vm_utils.ParamKeyCreate, createTime)
	msg.SetStringArray(vm_utils.ParamKeyInternal, internal)
	msg.SetStringArray(vm_utils.ParamKeyExternal, external)
	msg.SetStringArray(vm_utils.ParamKeyHardware, hardware)

	msg.SetStringArray(vm_utils.ParamKeyGroup, groups)
	msg.SetUIntArray(vm_utils.ParamKeyCore, cores)
	msg.SetUIntArray(vm_utils.ParamKeyOption, options)
	msg.SetUIntArray(vm_utils.ParamKeyEnable, enables)
	msg.SetUIntArray(vm_utils.ParamKeyProgress, progress)
	msg.SetUIntArray(vm_utils.ParamKeyStatus, status)
	msg.SetUIntArray(vm_utils.ParamKeyMemory, memories)
	msg.SetUIntArray(vm_utils.ParamKeyCount, diskCounts)
	msg.SetUIntArray(vm_utils.ParamKeyDisk, disks)
	msg.SetUIntArray(vm_utils.ParamKeyMedia, mediaAttached)
	msg.SetUIntArray(vm_utils.ParamKeyPriority, cpuPriorities)
	msg.SetUIntArray(vm_utils.ParamKeyLimit, ioLimits)
	return nil
}

func (config *InstanceStatus) Marshal(msg vm_utils.Message) error {
	msg.SetUInt(vm_utils.ParamKeyCore, config.Cores)
	msg.SetUInt(vm_utils.ParamKeyMemory, config.Memory)
	msg.SetUIntArray(vm_utils.ParamKeyDisk, config.Disks)

	msg.SetString(vm_utils.ParamKeyName, config.Name)
	msg.SetString(vm_utils.ParamKeyUser, config.User)
	msg.SetString(vm_utils.ParamKeyGroup, config.Group)
	msg.SetString(vm_utils.ParamKeyPool, config.Pool)
	msg.SetString(vm_utils.ParamKeyCell, config.Cell)
	msg.SetString(vm_utils.ParamKeyHost, config.Host)
	if config.ID != "" {
		msg.SetString(vm_utils.ParamKeyInstance, config.ID)
	}
	msg.SetBoolean(vm_utils.ParamKeyEnable, config.Created)
	msg.SetUInt(vm_utils.ParamKeyProgress, config.Progress)

	if config.AutoStart {
		msg.SetUIntArray(vm_utils.ParamKeyOption, []uint64{1})
	} else {
		msg.SetUIntArray(vm_utils.ParamKeyOption, []uint64{0})
	}
	msg.SetBoolean(vm_utils.ParamKeyMedia, config.MediaAttached)
	var insStatus uint
	if config.Running {
		insStatus = InstanceStatusRunning
	} else {
		insStatus = InstanceStatusStopped
	}
	if config.Lost {
		insStatus |= 1 << InstanceStatusLostBit
	}

	msg.SetUInt(vm_utils.ParamKeyStatus, insStatus)
	msg.SetString(vm_utils.ParamKeySecret, config.MonitorSecret)
	msg.SetString(vm_utils.ParamKeySystem, config.System)
	msg.SetString(vm_utils.ParamKeyCreate, config.CreateTime)
	msg.SetString(vm_utils.ParamKeyHardware, config.HardwareAddress)
	var internalMonitor = fmt.Sprintf("%s:%d", config.InternalNetwork.MonitorAddress, config.InternalNetwork.MonitorPort)
	var externalMonitor = fmt.Sprintf("%s:%d", config.ExternalNetwork.MonitorAddress, config.ExternalNetwork.MonitorPort)
	msg.SetStringArray(vm_utils.ParamKeyMonitor, []string{internalMonitor, externalMonitor})
	msg.SetStringArray(vm_utils.ParamKeyAddress, []string{config.InternalNetwork.InstanceAddress, config.ExternalNetwork.InstanceAddress})
	msg.SetString(vm_utils.ParamKeyInternal, config.InternalNetwork.AssignedAddress)
	msg.SetString(vm_utils.ParamKeyExternal, config.ExternalNetwork.AssignedAddress)
	//QoS
	msg.SetUInt(vm_utils.ParamKeyPriority, uint(config.CPUPriority))
	msg.SetUIntArray(vm_utils.ParamKeyLimit, []uint64{config.ReadSpeed, config.WriteSpeed, config.ReadIOPS,
		config.WriteIOPS, config.ReceiveSpeed, config.SendSpeed})
	return nil
}
