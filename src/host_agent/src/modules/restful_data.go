package modules

import (
	"encoding/binary"
	"fmt"
	"vm_manager/vm_utils"

	"net"
)

type restAddressList struct {
	NetworkAddress   string `json:"network_address,omitempty"`
	DisplayAddress   string `json:"display_address,omitempty"`
	AllocatedAddress string `json:"allocated_address,omitempty"`
}

type restGuestConfig struct {
	Name            string          `json:"name"`
	ID              string          `json:"id,omitempty"`
	Created         bool            `json:"created"`
	Progress        uint            `json:"progress"`
	Running         bool            `json:"running"`
	Lost            bool            `json:"lost,omitempty"`
	Owner           string          `json:"owner"`
	Group           string          `json:"group"`
	Pool            string          `json:"pool,omitempty"`
	Cell            string          `json:"cell,omitempty"`
	Host            string          `json:"host,omitempty"`
	Cores           uint            `json:"cores"`
	Memory          uint            `json:"memory"`
	TotalDisk       uint64          `json:"total_disk"`
	Disks           []uint64        `json:"disks"`
	AutoStart       bool            `json:"auto_start"`
	System          string          `json:"system,omitempty"`
	MonitorSecret   string          `json:"monitor_secret,omitempty"`
	EthernetAddress string          `json:"ethernet_address,omitempty"`
	DisplayProtocol string          `json:"display_protocol,omitempty"`
	Internal        restAddressList `json:"internal,omitempty"`
	External        restAddressList `json:"external,omitempty"`
	CreateTime      string          `json:"create_time,omitempty"`
	MediaAttached   bool            `json:"media_attached,omitempty"`
	QoS             restInstanceQoS `json:"qos,omitempty"`
}

type restInstanceStatus struct {
	restGuestConfig
	MediaAttached   bool    `json:"media_attached,omitemtpy"`
	MediaSource     string  `json:"media_source,omitemtpy"`
	CpuUsage        float64 `json:"cpu_usage"`
	MemoryAvailable uint64  `json:"memory_available"`
	DiskAvailable   uint64  `json:"disk_available"`
	BytesRead       uint64  `json:"bytes_read"`
	BytesWritten    uint64  `json:"bytes_written"`
	BytesReceived   uint64  `json:"bytes_received"`
	BytesSent       uint64  `json:"bytes_sent"`
}

type restInstanceQoS struct {
	CPUPriority  string `json:"cpu_priority,omitempty"`
	WriteSpeed   uint64 `json:"write_speed,omitempty"`
	WriteIOPS    uint64 `json:"write_iops,omitempty"`
	ReadSpeed    uint64 `json:"read_speed,omitempty"`
	ReadIOPS     uint64 `json:"read_iops,omitempty"`
	ReceiveSpeed uint64 `json:"receive_speed,omitempty"`
	SendSpeed    uint64 `json:"send_speed,omitempty"`
}

const (
	priority_label_high   = "high"
	priority_label_medium = "medium"
	priority_label_low    = "low"
)

func UnmarshalGuestConfigListFromMessage(msg vm_utils.Message) (result []restGuestConfig, err error) {
	result = make([]restGuestConfig, 0)

	count, err := msg.GetUInt(vm_utils.ParamKeyCount)
	if err != nil {
		return result, err
	}
	var names, ids, pools, cells, hosts, users, groups, monitors, addresses, systems,
		createTime, internal, external, hardware []string
	var cores, options, enables, progress, status, memories, disks, diskCounts, mediaAttached, cpuPriorities, ioLimits []uint64
	if pools, err = msg.GetStringArray(vm_utils.ParamKeyPool); err != nil {
		return result, err
	}
	if cells, err = msg.GetStringArray(vm_utils.ParamKeyCell); err != nil {
		return result, err
	}
	if hosts, err = msg.GetStringArray(vm_utils.ParamKeyHost); nil == err {
		if int(count) != len(hosts) {
			err = fmt.Errorf("unexpected hosts count %d / %d", len(hosts), count)
			return
		}
	}
	if names, err = msg.GetStringArray(vm_utils.ParamKeyName); err != nil {
		return result, err
	}
	if ids, err = msg.GetStringArray(vm_utils.ParamKeyInstance); err != nil {
		return result, err
	}
	if users, err = msg.GetStringArray(vm_utils.ParamKeyUser); err != nil {
		return result, err
	}
	if groups, err = msg.GetStringArray(vm_utils.ParamKeyGroup); err != nil {
		return result, err
	}
	if monitors, err = msg.GetStringArray(vm_utils.ParamKeyMonitor); err != nil {
		return result, err
	}
	if addresses, err = msg.GetStringArray(vm_utils.ParamKeyAddress); err != nil {
		return result, err
	}
	if internal, err = msg.GetStringArray(vm_utils.ParamKeyInternal); err != nil {
		return result, err
	}
	if external, err = msg.GetStringArray(vm_utils.ParamKeyExternal); err != nil {
		return result, err
	}
	if hardware, err = msg.GetStringArray(vm_utils.ParamKeyHardware); err != nil {
		return
	}
	if cores, err = msg.GetUIntArray(vm_utils.ParamKeyCore); err != nil {
		return result, err
	}
	if options, err = msg.GetUIntArray(vm_utils.ParamKeyOption); err != nil {
		return result, err
	}
	if enables, err = msg.GetUIntArray(vm_utils.ParamKeyEnable); err != nil {
		return result, err
	}
	if progress, err = msg.GetUIntArray(vm_utils.ParamKeyProgress); err != nil {
		return result, err
	}
	if status, err = msg.GetUIntArray(vm_utils.ParamKeyStatus); err != nil {
		return result, err
	}
	if memories, err = msg.GetUIntArray(vm_utils.ParamKeyMemory); err != nil {
		return result, err
	}
	if diskCounts, err = msg.GetUIntArray(vm_utils.ParamKeyCount); err != nil {
		return result, err
	}
	if disks, err = msg.GetUIntArray(vm_utils.ParamKeyDisk); err != nil {
		return result, err
	}
	if mediaAttached, err = msg.GetUIntArray(vm_utils.ParamKeyMedia); err != nil {
		return result, err
	}
	if systems, err = msg.GetStringArray(vm_utils.ParamKeySystem); err != nil {
		return
	}

	if createTime, err = msg.GetStringArray(vm_utils.ParamKeyCreate); err != nil {
		return
	}

	if cpuPriorities, err = msg.GetUIntArray(vm_utils.ParamKeyPriority); err != nil {
		return
	}

	if ioLimits, err = msg.GetUIntArray(vm_utils.ParamKeyLimit); err != nil {
		return
	}

	const (
		ReadSpeedOffset = iota
		WriteSpeedOffset
		ReadIOPSOffset
		WriteIOPSOffset
		ReceiveOffset
		SendOffset
		ValidLimitParametersCount = 6
	)
	if len(cpuPriorities) != int(count) {
		err = fmt.Errorf("unpected priority array size %d / %d", len(cpuPriorities), count)
		return
	}

	if len(ioLimits) != int(count*ValidLimitParametersCount) {
		err = fmt.Errorf("unpected limit array size %d / %d", len(ioLimits), count*ValidLimitParametersCount)
		return
	}

	var diskOffset = 0
	for i := 0; i < int(count); i++ {
		var config restGuestConfig
		config.Pool = pools[i]
		config.Cell = cells[i]
		if 0 != len(hosts) {
			config.Host = hosts[i]
		}
		config.Name = names[i]
		config.ID = ids[i]
		config.Owner = users[i]
		config.Group = groups[i]
		config.Internal.DisplayAddress = monitors[i*2]
		config.External.DisplayAddress = monitors[i*2+1]
		config.Internal.NetworkAddress = addresses[i*2]
		config.External.NetworkAddress = addresses[i*2+1]
		config.Internal.AllocatedAddress = internal[i]
		config.External.AllocatedAddress = external[i]

		config.Cores = uint(cores[i])
		config.Memory = uint(memories[i])
		var diskCount = int(diskCounts[i])
		for _, diskSize := range disks[diskOffset : diskOffset+diskCount] {
			config.Disks = append(config.Disks, diskSize)
			config.TotalDisk += diskSize
		}
		diskOffset += diskCount
		config.Progress = uint(progress[i])
		config.AutoStart = options[i] == 1
		config.Created = enables[i] == 1
		config.MediaAttached = mediaAttached[i] == 1
		if 0 != (status[i] >> InstanceStatusLostBit) {
			config.Lost = true
		}
		var mask = uint64(1<<InstanceStatusLostBit - 1)
		if InstanceStatusRunning == (status[i] & mask) {
			config.Running = true
		} else {
			config.Running = false
		}
		config.System = systems[i]
		config.CreateTime = createTime[i]
		config.EthernetAddress = hardware[i]
		switch PriorityEnum(cpuPriorities[i]) {
		case PriorityHigh:
			config.QoS = restInstanceQoS{CPUPriority: priority_label_high}
		case PriorityMedium:
			config.QoS = restInstanceQoS{CPUPriority: priority_label_medium}
		case PriorityLow:
			config.QoS = restInstanceQoS{CPUPriority: priority_label_low}
		default:
			err = fmt.Errorf("invalid CPU priority %d", cpuPriorities[i])
			return
		}
		config.QoS.ReadSpeed = ioLimits[ValidLimitParametersCount*i+ReadSpeedOffset]
		config.QoS.WriteSpeed = ioLimits[ValidLimitParametersCount*i+WriteSpeedOffset]
		config.QoS.ReadIOPS = ioLimits[ValidLimitParametersCount*i+ReadIOPSOffset]
		config.QoS.WriteIOPS = ioLimits[ValidLimitParametersCount*i+WriteIOPSOffset]
		config.QoS.ReceiveSpeed = ioLimits[ValidLimitParametersCount*i+ReceiveOffset]
		config.QoS.SendSpeed = ioLimits[ValidLimitParametersCount*i+SendOffset]
		result = append(result, config)
	}
	return result, nil
}

func (config *restGuestConfig) Unmarshal(msg vm_utils.Message) (err error) {
	//require fields
	if config.Name, err = msg.GetString(vm_utils.ParamKeyName); err != nil {
		return err
	}
	if config.Created, err = msg.GetBoolean(vm_utils.ParamKeyEnable); err != nil {
		return err
	}
	if status, err := msg.GetUInt(vm_utils.ParamKeyStatus); err != nil {
		return err
	} else {
		if 0 != (status >> InstanceStatusLostBit) {
			config.Lost = true
		}
		var mask = uint(1<<InstanceStatusLostBit - 1)
		if InstanceStatusRunning == (status & mask) {
			config.Running = true
		} else {
			config.Running = false
		}
	}
	options, err := msg.GetUIntArray(vm_utils.ParamKeyOption)
	if err != nil {
		return err
	}
	const (
		ValidOptionsCount       = 1
		ValidNetworkParamsCount = 2
	)
	if ValidOptionsCount != len(options) {
		return fmt.Errorf("unexpected options params count %d", len(options))
	}
	config.AutoStart = 1 == options[0]
	if config.Owner, err = msg.GetString(vm_utils.ParamKeyUser); err != nil {
		return err
	}
	if config.Group, err = msg.GetString(vm_utils.ParamKeyGroup); err != nil {
		return err
	}
	if config.Cores, err = msg.GetUInt(vm_utils.ParamKeyCore); err != nil {
		return err
	}
	if config.Memory, err = msg.GetUInt(vm_utils.ParamKeyMemory); err != nil {
		return err
	}
	if config.Disks, err = msg.GetUIntArray(vm_utils.ParamKeyDisk); err != nil {
		return err
	}
	if config.MediaAttached, err = msg.GetBoolean(vm_utils.ParamKeyMedia); err != nil {
		return err
	}
	config.TotalDisk = 0
	for _, size := range config.Disks {
		config.TotalDisk += size
	}
	if config.MonitorSecret, err = msg.GetString(vm_utils.ParamKeySecret); err != nil {
		return err
	}
	if addresses, err := msg.GetStringArray(vm_utils.ParamKeyAddress); err != nil {
		return err
	} else if len(addresses) != ValidNetworkParamsCount {
		return fmt.Errorf("expected address params cound %d", len(addresses))
	} else {
		config.Internal.NetworkAddress = addresses[0]
		config.External.NetworkAddress = addresses[1]
	}

	if monitor, err := msg.GetStringArray(vm_utils.ParamKeyMonitor); err != nil {
		return err
	} else if len(monitor) != ValidNetworkParamsCount {
		return fmt.Errorf("expected monitor params cound %d", len(monitor))
	} else {
		config.Internal.DisplayAddress = monitor[0]
		config.External.DisplayAddress = monitor[1]
	}
	if internal, err := msg.GetString(vm_utils.ParamKeyInternal); err == nil {
		config.Internal.AllocatedAddress = internal
	}

	if external, err := msg.GetString(vm_utils.ParamKeyExternal); err == nil {
		config.External.AllocatedAddress = external
	}

	if system, err := msg.GetString(vm_utils.ParamKeySystem); err == nil {
		config.System = system
	}
	if createTime, err := msg.GetString(vm_utils.ParamKeyCreate); err == nil {
		config.CreateTime = createTime
	}
	if hardware, err := msg.GetString(vm_utils.ParamKeyHardware); err == nil {
		config.EthernetAddress = hardware
	}
	if id, err := msg.GetString(vm_utils.ParamKeyInstance); err == nil {
		config.ID = id
	}
	if progress, err := msg.GetUInt(vm_utils.ParamKeyProgress); err == nil {
		config.Progress = progress
	}
	if pool, err := msg.GetString(vm_utils.ParamKeyPool); err == nil {
		config.Pool = pool
	}
	if cell, err := msg.GetString(vm_utils.ParamKeyCell); err == nil {
		config.Cell = cell
	}
	priorityValue, _ := msg.GetUInt(vm_utils.ParamKeyPriority)
	switch PriorityEnum(priorityValue) {
	case PriorityHigh:
		config.QoS = restInstanceQoS{CPUPriority: priority_label_high}
	case PriorityMedium:
		config.QoS = restInstanceQoS{CPUPriority: priority_label_medium}
	case PriorityLow:
		config.QoS = restInstanceQoS{CPUPriority: priority_label_low}
	default:
		err = fmt.Errorf("invalid CPU priority %d", priorityValue)
		return
	}

	if limitParameters, err := msg.GetUIntArray(vm_utils.ParamKeyLimit); err == nil {
		const (
			ReadSpeedOffset = iota
			WriteSpeedOffset
			ReadIOPSOffset
			WriteIOPSOffset
			ReceiveOffset
			SendOffset
			ValidLimitParametersCount = 6
		)

		if ValidLimitParametersCount != len(limitParameters) {
			err = fmt.Errorf("invalid QoS parameters count %d", len(limitParameters))
			return err
		}
		config.QoS.ReadSpeed = limitParameters[ReadSpeedOffset]
		config.QoS.WriteSpeed = limitParameters[WriteSpeedOffset]
		config.QoS.ReadIOPS = limitParameters[ReadIOPSOffset]
		config.QoS.WriteIOPS = limitParameters[WriteIOPSOffset]
		config.QoS.ReceiveSpeed = limitParameters[ReceiveOffset]
		config.QoS.SendSpeed = limitParameters[SendOffset]
	}

	return nil
}

func (status *restInstanceStatus) Unmarshal(msg vm_utils.Message) (err error) {
	if err = status.restGuestConfig.Unmarshal(msg); err != nil {
		return err
	}
	if status.CpuUsage, err = msg.GetFloat(vm_utils.ParamKeyUsage); err != nil {
		return err
	}
	const (
		ValidAvailableParams = 2
		ValidIOParams        = 4
	)
	if available, err := msg.GetUIntArray(vm_utils.ParamKeyAvailable); err != nil {
		return err
	} else if ValidAvailableParams != len(available) {
		return fmt.Errorf("invalid available params count %d", len(available))
	} else {
		status.MemoryAvailable = available[0]
		status.DiskAvailable = available[1]
	}
	if ios, err := msg.GetUIntArray(vm_utils.ParamKeyIO); err != nil {
		return err
	} else if ValidIOParams != len(ios) {
		return fmt.Errorf("invalid io params count %d", len(ios))
	} else {
		status.BytesRead = ios[0]
		status.BytesWritten = ios[1]
		status.BytesReceived = ios[2]
		status.BytesSent = ios[3]
	}
	if attached, err := msg.GetBoolean(vm_utils.ParamKeyMedia); nil == err {
		status.MediaAttached = attached
	}
	if source, err := msg.GetString(vm_utils.ParamKeySource); nil == err {
		status.MediaSource = source
	}
	return nil
}

type restSecurityPolicyRule struct {
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	FromAddress string `json:"from_address,omitempty"`
	ToAddress   string `json:"to_address,omitempty"`
	ToPort      uint   `json:"to_port"`
}

type restSecurityPolicyGroup struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	User          string `json:"user"`
	Group         string `json:"group"`
	Enabled       bool   `json:"enabled"`
	Global        bool   `json:"global"`
	DefaultAction string `json:"default_action"`
}

type restGuestSecurityPolicy struct {
	DefaultAction string                   `json:"default_action"`
	Rules         []restSecurityPolicyRule `json:"rules,omitempty"`
}

const (
	actionStringAccept = "accept"
	actionStringReject = "reject"
)

func (rule *restSecurityPolicyRule) build(msg vm_utils.Message) {
	msg.SetString(vm_utils.ParamKeyProtocol, rule.Protocol)
	msg.SetString(vm_utils.ParamKeyFrom, rule.FromAddress)
	msg.SetString(vm_utils.ParamKeyTo, rule.ToAddress)
	msg.SetUInt(vm_utils.ParamKeyPort, rule.ToPort)
	if actionStringAccept == rule.Action {
		msg.SetBoolean(vm_utils.ParamKeyAction, true)
	} else {
		msg.SetBoolean(vm_utils.ParamKeyAction, false)
	}
}

func (rule *restSecurityPolicyRule) buildForCell(msg vm_utils.Message) (err error) {
	switch rule.Protocol {
	case PolicyRuleProtocolTCP:
		msg.SetUInt(vm_utils.ParamKeyProtocol, PolicyRuleProtocolIndexTCP)
	case PolicyRuleProtocolUDP:
		msg.SetUInt(vm_utils.ParamKeyProtocol, PolicyRuleProtocolIndexUDP)
	case PolicyRuleProtocolICMP:
		msg.SetUInt(vm_utils.ParamKeyProtocol, PolicyRuleProtocolIndexICMP)
	default:
		err = fmt.Errorf("invalid protocol '%s'", rule.Protocol)
		return
	}
	msg.SetUInt(vm_utils.ParamKeyFrom, uint(IPv4ToUInt32(rule.FromAddress)))
	msg.SetUInt(vm_utils.ParamKeyTo, uint(IPv4ToUInt32(rule.ToAddress)))
	msg.SetUInt(vm_utils.ParamKeyPort, rule.ToPort)
	if actionStringAccept == rule.Action {
		msg.SetBoolean(vm_utils.ParamKeyAction, true)
	} else {
		msg.SetBoolean(vm_utils.ParamKeyAction, false)
	}
	return nil
}

func (policy *restSecurityPolicyGroup) build(msg vm_utils.Message) {
	if "" != policy.ID {
		msg.SetString(vm_utils.ParamKeyPolicy, policy.ID)
	}
	msg.SetString(vm_utils.ParamKeyName, policy.Name)
	msg.SetString(vm_utils.ParamKeyDescription, policy.Description)
	msg.SetString(vm_utils.ParamKeyUser, policy.User)
	msg.SetString(vm_utils.ParamKeyGroup, policy.Group)
	if actionStringAccept == policy.DefaultAction {
		msg.SetBoolean(vm_utils.ParamKeyAction, true)
	} else {
		msg.SetBoolean(vm_utils.ParamKeyAction, false)
	}

	msg.SetBoolean(vm_utils.ParamKeyEnable, policy.Enabled)
	msg.SetBoolean(vm_utils.ParamKeyLimit, policy.Global)
}

func (config *AddressPoolConfig) build(msg vm_utils.Message) (err error) {
	switch config.Provider {
	case AddressProviderDHCP:
	case AddressProviderCloudInit:
	default:
		err = fmt.Errorf("invalid provider '%s'", config.Provider)
		return
	}
	if "" != config.Mode {
		switch config.Mode {
		case AddressAllocationInternal:
		case AddressAllocationExternal:
		case AddressAllocationBoth:
		default:
			err = fmt.Errorf("invalid allocation mode '%s'", config.Mode)
			return
		}
	}
	msg.SetString(vm_utils.ParamKeyMode, config.Provider)
	//msg.SetString(vm_utils.ParamKeyAllocate, config.Mode)
	msg.SetString(vm_utils.ParamKeyAddress, config.Name)
	msg.SetString(vm_utils.ParamKeyGateway, config.Gateway)
	msg.SetStringArray(vm_utils.ParamKeyServer, config.DNS)
	return nil
}

func parsePolicyRuleList(msg vm_utils.Message) (rules []restSecurityPolicyRule, err error) {
	rules = make([]restSecurityPolicyRule, 0)
	var from, protocol []string
	var actions, ports []uint64
	if from, err = msg.GetStringArray(vm_utils.ParamKeyFrom); err != nil {
		err = fmt.Errorf("get source address fail: %s", err.Error())
		return
	}
	var elementCount = len(from)
	if protocol, err = msg.GetStringArray(vm_utils.ParamKeyProtocol); err != nil {
		err = fmt.Errorf("get protocol fail: %s", err.Error())
		return
	} else if len(protocol) != elementCount {
		err = fmt.Errorf("invalid protocol count %d", len(protocol))
		return
	}
	if actions, err = msg.GetUIntArray(vm_utils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get action fail: %s", err.Error())
		return
	} else if len(actions) != elementCount {
		err = fmt.Errorf("invalid action count %d", len(actions))
		return
	}
	if ports, err = msg.GetUIntArray(vm_utils.ParamKeyPort); err != nil {
		err = fmt.Errorf("get target port fail: %s", err.Error())
		return
	} else if len(ports) != elementCount {
		err = fmt.Errorf("invalid target port count %d", len(ports))
		return
	}
	for i := 0; i < elementCount; i++ {
		var rule = restSecurityPolicyRule{
			FromAddress: from[i],
			ToPort:      uint(ports[i]),
			Protocol:    protocol[i],
		}
		if PolicyRuleActionAccept == actions[i] {
			rule.Action = actionStringAccept
		} else {
			rule.Action = actionStringReject
		}
		rules = append(rules, rule)
	}
	return
}

func parsePolicyGroupList(msg vm_utils.Message) (groups []restSecurityPolicyGroup, err error) {
	const (
		flagFalse = iota
		flagTrue
	)
	groups = make([]restSecurityPolicyGroup, 0)
	var id, name, description, user, group []string
	var action, enabled, global []uint64
	if id, err = msg.GetStringArray(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy id fail: %s", err.Error())
		return
	}
	var elementCount = len(id)
	if name, err = msg.GetStringArray(vm_utils.ParamKeyName); err != nil {
		err = fmt.Errorf("get policy name fail: %s", err.Error())
		return
	} else if len(name) != elementCount {
		err = fmt.Errorf("invalid name count %d", len(name))
		return
	}
	if description, err = msg.GetStringArray(vm_utils.ParamKeyDescription); err != nil {
		err = fmt.Errorf("get description fail: %s", err.Error())
		return
	} else if len(description) != elementCount {
		err = fmt.Errorf("invalid description count %d", len(description))
		return
	}
	if user, err = msg.GetStringArray(vm_utils.ParamKeyUser); err != nil {
		err = fmt.Errorf("get user fail: %s", err.Error())
		return
	} else if len(user) != elementCount {
		err = fmt.Errorf("invalid user count %d", len(user))
		return
	}
	if group, err = msg.GetStringArray(vm_utils.ParamKeyGroup); err != nil {
		err = fmt.Errorf("get group fail: %s", err.Error())
		return
	} else if len(group) != elementCount {
		err = fmt.Errorf("invalid group count %d", len(group))
		return
	}
	if action, err = msg.GetUIntArray(vm_utils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get action fail: %s", err.Error())
		return
	} else if len(action) != elementCount {
		err = fmt.Errorf("invalid action count %d", len(action))
		return
	}
	if enabled, err = msg.GetUIntArray(vm_utils.ParamKeyEnable); err != nil {
		err = fmt.Errorf("get enable flag fail: %s", err.Error())
		return
	} else if len(enabled) != elementCount {
		err = fmt.Errorf("invalid enable flag count %d", len(enabled))
		return
	}
	if global, err = msg.GetUIntArray(vm_utils.ParamKeyLimit); err != nil {
		err = fmt.Errorf("get global flag fail: %s", err.Error())
		return
	} else if len(global) != elementCount {
		err = fmt.Errorf("invalid global flag count %d", len(global))
		return
	}
	for i := 0; i < elementCount; i++ {
		var policy = restSecurityPolicyGroup{
			ID:          id[i],
			Name:        name[i],
			User:        user[i],
			Group:       group[i],
			Description: description[i],
			Enabled:     flagTrue == enabled[i],
			Global:      flagTrue == global[i],
		}
		if PolicyRuleActionAccept == action[i] {
			policy.DefaultAction = actionStringAccept
		} else {
			policy.DefaultAction = actionStringReject
		}
		groups = append(groups, policy)
	}
	return
}

func parsePolicyGroup(msg vm_utils.Message) (policy restSecurityPolicyGroup, err error) {
	if policy.ID, err = msg.GetString(vm_utils.ParamKeyPolicy); err != nil {
		err = fmt.Errorf("get policy id fail: %s", err.Error())
		return
	}
	if policy.Name, err = msg.GetString(vm_utils.ParamKeyName); err != nil {
		err = fmt.Errorf("get name fail: %s", err.Error())
		return
	}
	if policy.Description, err = msg.GetString(vm_utils.ParamKeyDescription); err != nil {
		err = fmt.Errorf("get description fail: %s", err.Error())
		return
	}
	if policy.User, err = msg.GetString(vm_utils.ParamKeyUser); err != nil {
		err = fmt.Errorf("get user fail: %s", err.Error())
		return
	}
	if policy.Group, err = msg.GetString(vm_utils.ParamKeyGroup); err != nil {
		err = fmt.Errorf("get group fail: %s", err.Error())
		return
	}
	var accept bool
	if accept, err = msg.GetBoolean(vm_utils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get accept flag fail: %s", err.Error())
		return
	}
	if accept {
		policy.DefaultAction = actionStringAccept
	} else {
		policy.DefaultAction = actionStringReject
	}
	if policy.Enabled, err = msg.GetBoolean(vm_utils.ParamKeyEnable); err != nil {
		err = fmt.Errorf("get enabled flag fail: %s", err.Error())
		return
	}
	if policy.Global, err = msg.GetBoolean(vm_utils.ParamKeyLimit); err != nil {
		err = fmt.Errorf("get global flag fail: %s", err.Error())
		return
	}
	return
}

func parseGuestSecurityPolicy(msg vm_utils.Message) (policy restGuestSecurityPolicy, err error) {
	var from, to, protocol, actions, ports []uint64
	if from, err = msg.GetUIntArray(vm_utils.ParamKeyFrom); err != nil {
		err = fmt.Errorf("get source address fail: %s", err.Error())
		return
	}
	var elementCount = len(from)
	if to, err = msg.GetUIntArray(vm_utils.ParamKeyTo); err != nil {
		err = fmt.Errorf("get target address fail: %s", err.Error())
		return
	} else if len(to) != elementCount {
		err = fmt.Errorf("invalid target address count %d", len(to))
		return
	}
	if protocol, err = msg.GetUIntArray(vm_utils.ParamKeyProtocol); err != nil {
		err = fmt.Errorf("get protocol fail: %s", err.Error())
		return
	} else if len(protocol) != elementCount {
		err = fmt.Errorf("invalid protocol count %d", len(protocol))
		return
	}
	if actions, err = msg.GetUIntArray(vm_utils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get action fail: %s", err.Error())
		return
	} else if len(actions) != elementCount+1 {
		err = fmt.Errorf("invalid action count %d", len(actions))
		return
	}
	if ports, err = msg.GetUIntArray(vm_utils.ParamKeyPort); err != nil {
		err = fmt.Errorf("get target port fail: %s", err.Error())
		return
	} else if len(ports) != elementCount {
		err = fmt.Errorf("invalid target port count %d", len(ports))
		return
	}
	if PolicyRuleActionAccept == actions[elementCount] {
		policy.DefaultAction = actionStringAccept
	} else {
		policy.DefaultAction = actionStringReject
	}
	for i := 0; i < elementCount; i++ {

		var rule = restSecurityPolicyRule{
			FromAddress: UInt32ToIPv4(uint32(from[i])),
			ToAddress:   UInt32ToIPv4(uint32(to[i])),
			ToPort:      uint(ports[i]),
		}
		if PolicyRuleActionAccept == actions[i] {
			rule.Action = actionStringAccept
		} else {
			rule.Action = actionStringReject
		}
		switch protocol[i] {
		case PolicyRuleProtocolIndexTCP:
			rule.Protocol = PolicyRuleProtocolTCP
		case PolicyRuleProtocolIndexUDP:
			rule.Protocol = PolicyRuleProtocolUDP
		case PolicyRuleProtocolIndexICMP:
			rule.Protocol = PolicyRuleProtocolICMP
		default:
			err = fmt.Errorf("invalid protocol %d in %dth rule", protocol[i], i)
			return
		}
		policy.Rules = append(policy.Rules, rule)
	}
	return
}

func UInt32ToIPv4(input uint32) string {
	if 0 == input {
		return ""
	}
	var bytes = make([]byte, net.IPv4len)
	binary.BigEndian.PutUint32(bytes, input)
	return net.IP(bytes).String()
}

func IPv4ToUInt32(input string) uint32 {
	if "" == input {
		return 0
	}
	var ip = net.ParseIP(input)
	return binary.BigEndian.Uint32(ip.To4())
}
