package task

import (
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type AddComputePoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *AddComputePoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request add cell '%s' to pool '%s' from %s.[%08X]", id, cellName, poolName,
		request.GetSender(), request.GetFromSession())
	var respChan = make(chan error)
	executor.ResourceModule.AddCell(poolName, cellName, respChan)

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.AddComputePoolCellResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	defer executor.Sender.SendMessage(resp, request.GetSender())

	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] add compute cell fail: %s", id, err.Error())
		return
	}

	var computePool modules.ComputePoolInfo
	{
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.GetComputePool(poolName, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			resp.SetError(err.Error())
			log.Printf("[%08X] warning: get compute pool fail: %s", id, err.Error())
			return
		}
		computePool = result.ComputePoolConfig
	}

	{
		notify, _ := vm_utils.CreateJsonMessage(vm_utils.ComputePoolReadyEvent)
		notify.SetString(vm_utils.ParamKeyPool, poolName)
		notify.SetString(vm_utils.ParamKeyStorage, computePool.Storage)
		notify.SetString(vm_utils.ParamKeyNetwork, computePool.Network)
		if "" != computePool.Storage {
			var respChan = make(chan modules.ResourceResult, 1)
			executor.ResourceModule.GetStoragePool(computePool.Storage, respChan)
			var result = <-respChan
			if result.Error != nil {
				err = result.Error
				resp.SetError(err.Error())
				log.Printf("[%08X] get storage pool fail: %s", id, err.Error())
				return
			}
			var storagePool = result.StoragePool
			notify.SetString(vm_utils.ParamKeyType, storagePool.Type)
			notify.SetString(vm_utils.ParamKeyHost, storagePool.Host)
			notify.SetString(vm_utils.ParamKeyTarget, storagePool.Target)
		}
		if "" != computePool.Network {
			var respChan = make(chan modules.ResourceResult, 1)
			executor.ResourceModule.GetAddressPool(computePool.Network, respChan)
			var result = <-respChan
			if result.Error != nil {
				log.Printf("[%08X] get address pool fail: %s", id, result.Error.Error())
				return nil
			}
			var addressPool = result.AddressPool
			notify.SetString(vm_utils.ParamKeyGateway, addressPool.Gateway)
			notify.SetStringArray(vm_utils.ParamKeyServer, addressPool.DNS)
		}

		notify.SetFromSession(id)
		if err = executor.Sender.SendMessage(notify, cellName); err != nil {
			log.Printf("[%08X] notify cell fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return
		}
		const (
			AddCellTimeout = 10 * time.Second
		)
		timer := time.NewTimer(AddCellTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.GetID() != vm_utils.ComputeCellReadyEvent {
				log.Printf("[%08X] unexpected message [%08X] from %s.[%08X]", id, cellResp.GetID(),
					cellResp.GetSender(), cellResp.GetFromSession())
				resp.SetError("unexpected message received")
				return
			}
			if !cellResp.IsSuccess() {
				log.Printf("[%08X] wait cell ready fail: %s", id, cellResp.GetError())
				resp.SetError(cellResp.GetError())
			} else {
				resp.SetSuccess(true)
				log.Printf("[%08X] cell ready with storage pool '%s'", id, computePool.Storage)
			}
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait cell ready timeout", id)
			resp.SetError("wait cell ready timeout")
			return
		}
	}
	return nil
}
