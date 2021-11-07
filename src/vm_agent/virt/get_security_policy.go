package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type GetSecurityPolicyExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
}

func (executor *GetSecurityPolicyExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		err = fmt.Errorf("get instance id fail: %s", err.Error())
		return
	}

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.GetGuestRuleResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var respChan = make(chan VmAgentSvc.InstanceResult, 1)
	executor.InstanceModule.GetSecurityPolicy(instanceID, respChan)
	var result = <-respChan
	if nil != result.Error {
		err = result.Error
		log.Printf("[%08X] get security policy of instance '%s' fail: %s",
			id, instanceID, err.Error())
		resp.SetError(err.Error())
	} else {
		var policy = result.Policy
		var fromIP, toIP, toPort, protocols, actions []uint64
		for index, rule := range policy.Rules {
			fromIP = append(fromIP, uint64(VmAgentSvc.IPv4ToUInt32(rule.SourceAddress)))
			toIP = append(toIP, uint64(VmAgentSvc.IPv4ToUInt32(rule.TargetAddress)))
			toPort = append(toPort, uint64(rule.TargetPort))
			switch rule.Protocol {
			case VmAgentSvc.PolicyRuleProtocolTCP:
				protocols = append(protocols, uint64(VmAgentSvc.PolicyRuleProtocolIndexTCP))
			case VmAgentSvc.PolicyRuleProtocolUDP:
				protocols = append(protocols, uint64(VmAgentSvc.PolicyRuleProtocolIndexUDP))
			case VmAgentSvc.PolicyRuleProtocolICMP:
				protocols = append(protocols, uint64(VmAgentSvc.PolicyRuleProtocolIndexICMP))
			default:
				log.Printf("[%08X] warning: invalid protocol %s on %dth security rule of instance '%s'",
					id, rule.Protocol, index, instanceID)
				continue
			}
			if rule.Accept {
				actions = append(actions, VmAgentSvc.PolicyRuleActionAccept)
			} else {
				actions = append(actions, VmAgentSvc.PolicyRuleActionReject)
			}
		}
		if policy.Accept {
			actions = append(actions, VmAgentSvc.PolicyRuleActionAccept)
			log.Printf("[%08X] %d security rule(s) available for instance '%s', accept by default",
				id, len(toPort), instanceID)
		} else {
			actions = append(actions, VmAgentSvc.PolicyRuleActionReject)
			log.Printf("[%08X] %d security rule(s) available for instance '%s', reject by default",
				id, len(toPort), instanceID)
		}
		resp.SetUIntArray(VmUtils.ParamKeyFrom, fromIP)
		resp.SetUIntArray(VmUtils.ParamKeyTo, toIP)
		resp.SetUIntArray(VmUtils.ParamKeyPort, toPort)
		resp.SetUIntArray(VmUtils.ParamKeyProtocol, protocols)
		resp.SetUIntArray(VmUtils.ParamKeyAction, actions)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
