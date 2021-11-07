package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type ModifySecurityRuleExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *ModifySecurityRuleExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	var index int
	var accept bool
	var fromIP, toIP, toPort, protocol uint
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance id fail: %s", err.Error())
		return
	}
	if index, err = request.GetInt(VmUtils.ParamKeyIndex); err != nil {
		err = fmt.Errorf("get rule index fail: %s", err.Error())
		return
	}
	if accept, err = request.GetBoolean(VmUtils.ParamKeyAction); err != nil {
		err = fmt.Errorf("get action fail: %s", err.Error())
		return
	}
	if fromIP, err = request.GetUInt(VmUtils.ParamKeyFrom); err != nil {
		err = fmt.Errorf("get source address fail: %s", err.Error())
		return
	}
	if toIP, err = request.GetUInt(VmUtils.ParamKeyTo); err != nil {
		err = fmt.Errorf("get target address fail: %s", err.Error())
		return
	}
	if toPort, err = request.GetUInt(VmUtils.ParamKeyPort); err != nil {
		err = fmt.Errorf("get target port fail: %s", err.Error())
		return
	} else if 0 == toPort || toPort > 0xFFFF {
		err = fmt.Errorf("invalid target port %d", toPort)
		return
	}
	if protocol, err = request.GetUInt(VmUtils.ParamKeyProtocol); err != nil {
		err = fmt.Errorf("get protocol fail: %s", err.Error())
		return
	}
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.ModifyGuestRuleResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var rule = VmAgentSvc.SecurityPolicyRule{
		Accept:     accept,
		TargetPort: toPort,
	}

	switch protocol {
	case VmAgentSvc.PolicyRuleProtocolIndexTCP:
		rule.Protocol = VmAgentSvc.PolicyRuleProtocolTCP
	case VmAgentSvc.PolicyRuleProtocolIndexUDP:
		rule.Protocol = VmAgentSvc.PolicyRuleProtocolUDP
	case VmAgentSvc.PolicyRuleProtocolIndexICMP:
		rule.Protocol = VmAgentSvc.PolicyRuleProtocolICMP
	default:
		err = fmt.Errorf("invalid protocol %d for security rule", protocol)
		return
	}
	rule.SourceAddress = VmAgentSvc.UInt32ToIPv4(uint32(fromIP))
	rule.TargetAddress = VmAgentSvc.UInt32ToIPv4(uint32(toIP))

	var respChan = make(chan error, 1)
	executor.InstanceModule.ModifySecurityPolicyRule(instanceID, index, rule, respChan)
	err = <-respChan
	if nil != err {
		log.Printf("[%08X] modify %dth security rule of instance '%s' fail: %s",
			id, index, instanceID, err.Error())
		resp.SetError(err.Error())
	} else {
		if accept {
			log.Printf("[%08X] %dth security rule of instance '%s' changed to accept protocol '%s' from '%s' to '%s:%d'",
				id, index, instanceID, rule.Protocol, rule.SourceAddress, rule.TargetAddress, rule.TargetPort)
		} else {
			log.Printf("[%08X] %dth security rule of instance '%s' changed to reject protocol '%s' from '%s' to '%s:%d'",
				id, index, instanceID, rule.Protocol, rule.SourceAddress, rule.TargetAddress, rule.TargetPort)
		}
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
