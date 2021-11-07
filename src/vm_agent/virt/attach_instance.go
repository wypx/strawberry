package virt

import (
	"log"
	"time"

	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type AttachInstanceExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
	NetworkModule  VmAgentSvc.NetworkModule
}

func (executor *AttachInstanceExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.AttachInstanceResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	isFailover, err := request.GetBoolean(VmUtils.ParamKeyImmediate)
	if err != nil {
		log.Printf("[%08X] recv attach instance request from %s.[%08X] but get failover flag fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var sourceCell string
	if isFailover {
		sourceCell, err = request.GetString(VmUtils.ParamKeyCell)
		if err != nil {
			log.Printf("[%08X] recv failover attach request from %s.[%08X] but get source cell fail: %s",
				id, request.GetSender(), request.GetFromSession(), err.Error())
			return err
		}
	}
	idList, err := request.GetStringArray(VmUtils.ParamKeyInstance)
	if err != nil {
		log.Printf("[%08X] recv attach instance request from %s.[%08X] but get target intance fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	log.Printf("[%08X] recv attach %d instance(s) request from %s.[%08X]", id, len(idList), request.GetSender(), request.GetFromSession())
	var networkResource map[string]VmAgentSvc.InstanceNetworkResource
	{
		var respChan = make(chan VmAgentSvc.InstanceResult, 1)
		executor.InstanceModule.GetNetworkResources(idList, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			resp.SetError(err.Error())
			log.Printf("[%08X] get network resource fail: %s", id, err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		networkResource = result.NetworkResources
	}
	{
		var respChan = make(chan VmAgentSvc.NetworkResult, 1)
		executor.NetworkModule.AttachInstances(networkResource, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			resp.SetError(err.Error())
			log.Printf("[%08X] attach network resource fail: %s", id, err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		networkResource = result.Resources
	}
	{
		var respChan = make(chan error, 1)
		executor.StorageModule.AttachVolumeGroup(idList, respChan)
		err = <-respChan
		if err != nil {
			resp.SetError(err.Error())
			log.Printf("[%08X] attach storage resource fail: %s", id, err.Error())
			executor.detachResource(id, idList, true, false, false)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		var respChan = make(chan error, 1)
		executor.InstanceModule.AttachInstances(networkResource, respChan)
		err = <-respChan
		if err != nil {
			resp.SetError(err.Error())
			log.Printf("[%08X] attach instance resource fail: %s", id, err.Error())
			executor.detachResource(id, idList, true, true, false)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	log.Printf("[%08X] instance(s) attached", id)

	idList = idList[:0]
	var monitorPorts []uint64
	for instanceID, resource := range networkResource {
		idList = append(idList, instanceID)
		monitorPorts = append(monitorPorts, uint64(resource.MonitorPort))
	}

	if isFailover {

		//notify migrate finish
		var respChan = make(chan error, 1)
		executor.InstanceModule.MigrateInstances(idList, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] migrate instaince fail: %s", id, err.Error())
			executor.detachResource(id, idList, true, true, true)
			return nil
		}

		notify, _ := VmUtils.CreateJsonMessage(VmUtils.InstanceMigratedEvent)
		notify.SetSuccess(true)
		notify.SetFromSession(id)
		notify.SetStringArray(VmUtils.ParamKeyInstance, idList)
		notify.SetUIntArray(VmUtils.ParamKeyMonitor, monitorPorts)
		notify.SetBoolean(VmUtils.ParamKeyImmediate, true)
		notify.SetString(VmUtils.ParamKeyCell, sourceCell)
		if err = executor.Sender.SendMessage(notify, request.GetSender()); err != nil {
			log.Printf("[%08X] warning: notify migrate finish fail: %s", id, err.Error())
		}
		log.Printf("[%08X] %d instance(s) migrated success when failover", id, len(idList))
		return nil

	} else {
		resp.SetSuccess(true)
		log.Printf("[%08X] instance(s) attached", id)
		if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
			log.Printf("[%08X] warning: send attach response fail: %s", id, err.Error())
		}
		//wait migrate
		timer := time.NewTimer(VmAgentSvc.DefaultOperateTimeout)
		select {
		case migrateRequest := <-incoming:
			if migrateRequest.GetID() != VmUtils.MigrateInstanceRequest {
				//detach fail
				log.Printf("[%08X] unexpect message received from %s when wait migrate: %d", id, migrateRequest.GetSender(), migrateRequest.GetID())
				executor.detachResource(id, idList, true, true, true)
				return nil
			}
			var migrationID string
			if migrationID, err = migrateRequest.GetString(VmUtils.ParamKeyMigration); err != nil {
				log.Printf("[%08X] parse migration ID from %s fail: %d", id, migrateRequest.GetSender(), err.Error())
				executor.detachResource(id, idList, true, true, true)
				return nil
			}
			//invoke migrate
			var respChan = make(chan error, 1)
			executor.InstanceModule.MigrateInstances(idList, respChan)
			err = <-respChan
			if err != nil {
				log.Printf("[%08X] migrate instaince fail: %s", id, err.Error())
				executor.detachResource(id, idList, true, true, true)
				return nil
			}

			notify, _ := VmUtils.CreateJsonMessage(VmUtils.InstanceMigratedEvent)
			notify.SetSuccess(true)
			notify.SetFromSession(id)
			notify.SetStringArray(VmUtils.ParamKeyInstance, idList)
			notify.SetUIntArray(VmUtils.ParamKeyMonitor, monitorPorts)
			notify.SetString(VmUtils.ParamKeyMigration, migrationID)
			notify.SetBoolean(VmUtils.ParamKeyImmediate, false)
			if err = executor.Sender.SendMessage(notify, request.GetSender()); err != nil {
				log.Printf("[%08X] warning: notify migrate finish fail: %s", id, err.Error())
			}
			log.Printf("[%08X] %d instance(s) migrated success", id, len(idList))
			return nil

		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait migrate request timeout", id)
			executor.detachResource(id, idList, true, true, true)
			return nil
		}
	}
}

func (executor *AttachInstanceExecutor) detachResource(id VmUtils.SessionID, instances []string, detachNetwork, detachVolume, detachInstance bool) {
	var respChan = make(chan error, 1)
	var err error
	if detachInstance {
		executor.InstanceModule.DetachInstances(instances, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] detach instance fail: %s", id, err.Error())
		}
	}
	if detachVolume {
		executor.StorageModule.DetachVolumeGroup(instances, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] detach volume fail: %s", id, err.Error())
		}
	}
	if detachNetwork {
		executor.NetworkModule.DetachInstances(instances, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] detach network fail: %s", id, err.Error())
		}
	}
}
