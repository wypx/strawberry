package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QuerySecurityPolicyGroupsExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QuerySecurityPolicyGroupsExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var condition modules.SecurityPolicyGroupQueryCondition
	if condition.User, err = request.GetString(vm_utils.ParamKeyUser); err != nil {
		err = fmt.Errorf("get user fail: %s", err.Error())
		return
	}
	if condition.Group, err = request.GetString(vm_utils.ParamKeyGroup); err != nil {
		err = fmt.Errorf("get group fail: %s", err.Error())
		return
	}
	if condition.EnabledOnly, err = request.GetBoolean(vm_utils.ParamKeyEnable); err != nil {
		err = fmt.Errorf("get enable flag fail: %s", err.Error())
		return
	}
	if condition.GlobalOnly, err = request.GetBoolean(vm_utils.ParamKeyLimit); err != nil {
		err = fmt.Errorf("get global flag fail: %s", err.Error())
		return
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryPolicyGroupResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetTransactionID(request.GetTransactionID())
	resp.SetSuccess(false)
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.QuerySecurityPolicyGroups(condition, respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] query security policy groups fail: %s",
			id, err.Error())
		resp.SetError(err.Error())
	} else {
		var id, name, description, user, group []string
		var accept, enabled, global []uint64
		const (
			flagFalse = iota
			flagTrue
		)
		for _, policy := range result.PolicyGroupList {
			id = append(id, policy.ID)
			name = append(name, policy.Name)
			description = append(description, policy.Description)
			user = append(user, policy.User)
			group = append(group, policy.Group)
			if policy.Accept {
				accept = append(accept, modules.PolicyRuleActionAccept)
			} else {
				accept = append(accept, modules.PolicyRuleActionReject)
			}
			if policy.Enabled {
				enabled = append(enabled, flagTrue)
			} else {
				enabled = append(enabled, flagFalse)
			}
			if policy.Global {
				global = append(global, flagTrue)
			} else {
				global = append(global, flagFalse)
			}
		}
		resp.SetStringArray(vm_utils.ParamKeyPolicy, id)
		resp.SetStringArray(vm_utils.ParamKeyName, name)
		resp.SetStringArray(vm_utils.ParamKeyDescription, description)
		resp.SetStringArray(vm_utils.ParamKeyUser, user)
		resp.SetStringArray(vm_utils.ParamKeyGroup, group)
		resp.SetUIntArray(vm_utils.ParamKeyAction, accept)
		resp.SetUIntArray(vm_utils.ParamKeyEnable, enabled)
		resp.SetUIntArray(vm_utils.ParamKeyLimit, global)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
