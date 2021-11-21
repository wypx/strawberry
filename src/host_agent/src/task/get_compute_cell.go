package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"

	"github.com/pkg/errors"
)

type GetComputeCellExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetComputeCellExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err != nil {
		return err
	}

	log.Printf("[%08X] get compute cell '%s.%s' from %s.[%08X]", id, poolName, cellName,
		request.GetSender(), request.GetFromSession())

	var respChan = make(chan modules.ResourceResult)

	executor.ResourceModule.GetComputeCellStatus(poolName, cellName, respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetComputePoolCellResponse)
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	resp.SetToSession(request.GetFromSession())
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] get compute cell fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var s = result.ComputeCell

	//assemble
	resp.SetString(vm_utils.ParamKeyName, s.Name)
	resp.SetString(vm_utils.ParamKeyAddress, s.Address)
	resp.SetBoolean(vm_utils.ParamKeyEnable, s.Enabled)
	resp.SetBoolean(vm_utils.ParamKeyStatus, s.Alive)
	if !s.Alive || !s.Enabled {
		resp.SetSuccess(true)
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	forward, _ := vm_utils.CreateJsonMessage(vm_utils.GetComputePoolCellRequest)
	forward.SetFromSession(id)
	forward.SetString(vm_utils.ParamKeyCell, cellName)
	if err = executor.Sender.SendMessage(forward, cellName); err != nil {
		log.Printf("[%08X] forward to cell '%s' fail: %s", id, cellName, err.Error())
		err = fmt.Errorf("forward to cell '%s' fail: %s", cellName, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var timer = time.NewTimer(modules.DefaultOperateTimeout)
	select {
	case <-timer.C:
		log.Printf("[%08X] wait cell status timeout", id)
		err = errors.New("wait cell status timeout")
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	case cellResp := <-incoming:
		if !cellResp.IsSuccess() {
			err = errors.New(cellResp.GetError())
			log.Printf("[%08X] get remote cell status fail: %s", id, err)
			resp.SetError(err.Error())
		} else {
			//success
			var storage, errors []string
			var attached []uint64
			storage, _ = cellResp.GetStringArray(vm_utils.ParamKeyStorage)
			errors, _ = cellResp.GetStringArray(vm_utils.ParamKeyError)
			attached, _ = cellResp.GetUIntArray(vm_utils.ParamKeyAttach)
			resp.SetStringArray(vm_utils.ParamKeyStorage, storage)
			resp.SetStringArray(vm_utils.ParamKeyError, errors)
			resp.SetUIntArray(vm_utils.ParamKeyAttach, attached)
			resp.SetSuccess(true)
		}
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
}
