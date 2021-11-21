package task

import (
	"errors"
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type StartBatchCreateGuestExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *StartBatchCreateGuestExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var poolName, namePrefix string
	var nameRule, guestCount uint
	if namePrefix, err = request.GetString(vm_utils.ParamKeyName); err != nil {
		return err
	}
	if poolName, err = request.GetString(vm_utils.ParamKeyPool); err != nil {
		return err
	}
	if nameRule, err = request.GetUInt(vm_utils.ParamKeyMode); err != nil {
		return err
	}
	if guestCount, err = request.GetUInt(vm_utils.ParamKeyCount); err != nil {
		return err
	}
	var templateID, adminName string
	var templateOptions []uint64
	if templateID, err = request.GetString(vm_utils.ParamKeyTemplate); err != nil {
		err = fmt.Errorf("get template id fail: %s", err.Error())
		return
	} else {
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.GetSystemTemplate(templateID, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = fmt.Errorf("get template fail: %s", result.Error)
			return
		}
		var t = result.Template
		if adminName, err = request.GetString(vm_utils.ParamKeyAdmin); err != nil {
			adminName = t.Admin
		}
		if templateOptions, err = t.ToOptions(); err != nil {
			err = fmt.Errorf("invalid template: %s", err.Error())
			return
		}
	}

	log.Printf("[%08X] recv batch create %d guests from %s.[%08X]", id, guestCount, request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.StartBatchCreateGuestResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)
	var originalSender = request.GetSender()

	var batchID string
	var guestList []string
	{
		var respChan = make(chan modules.ResourceResult, 1)
		var bathRequest = modules.BatchCreateRequest{modules.BatchCreateNameRule(nameRule), namePrefix, poolName, int(guestCount)}
		executor.ResourceModule.StartBatchCreateGuest(bathRequest, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] start batch create guest fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		batchID = result.Batch
		for _, status := range result.BatchCreate {
			guestList = append(guestList, status.Name)
		}
	}

	var targets = map[vm_utils.TransactionID]string{}
	//forward request
	for index, guestName := range guestList {
		var forward = vm_utils.CloneJsonMessage(request)
		var transID = vm_utils.TransactionID(index)
		forward.SetID(vm_utils.CreateGuestRequest)
		forward.SetString(vm_utils.ParamKeyAdmin, adminName)
		forward.SetString(vm_utils.ParamKeyName, guestName)
		forward.SetUIntArray(vm_utils.ParamKeyTemplate, templateOptions)
		forward.SetToSession(0)
		forward.SetFromSession(id)
		forward.SetTransactionID(transID)
		targets[transID] = guestName
		if err = executor.Sender.SendToSelf(forward); err != nil {
			log.Printf("[%08X] warning: forward create guest '%s' fail: %s", id, guestName, err.Error())
		}
	}
	log.Printf("[%08X] new batch create '%s' started", id, batchID)
	resp.SetSuccess(true)
	resp.SetString(vm_utils.ParamKeyID, batchID)
	executor.Sender.SendMessage(resp, originalSender)

	var lastUpdate = time.Now()
	const (
		CheckInterval = time.Second * 5
		UpdateTimeout = time.Second * 10
	)
	var checkTicker = time.NewTicker(CheckInterval)
	for len(targets) > 0 {
		select {
		case <-checkTicker.C:
			//check
			if lastUpdate.Add(UpdateTimeout).Before(time.Now()) {
				log.Printf("[%08X] warning: receive create response timeout", id)
				return
			}
		case createResponse := <-incoming:
			var transID = createResponse.GetTransactionID()
			guestName, exists := targets[transID]
			if !exists {
				log.Printf("[%08X] warning: invalid create response with trans [%08X] from [%08X]",
					id, transID, createResponse.GetFromSession())
				break
			}
			var errChan = make(chan error, 1)
			if createResponse.IsSuccess() {
				var guestID string
				if guestID, err = createResponse.GetString(vm_utils.ParamKeyInstance); err != nil {
					log.Printf("[%08X] warning: guest '%s' created, but get id fail", id, guestName)
					break
				}
				log.Printf("[%08X] create guest '%s'('%s') started", id, guestName, guestID)
				executor.ResourceModule.SetBatchCreateGuestStart(batchID, guestName, guestID, errChan)
			} else {
				var createError = errors.New(createResponse.GetError())
				log.Printf("[%08X] create guest '%s' fail: %s", id, guestName, createError.Error())
				executor.ResourceModule.SetBatchCreateGuestFail(batchID, guestName, createError, errChan)
			}
			var result = <-errChan
			if result != nil {
				log.Printf("[%08X] warning:update create status fail: %s", id, result.Error())
			}
			lastUpdate = time.Now()
			delete(targets, transID)
		}
	}
	//all targets finished
	log.Printf("[%08X] all create request finished in batch '%s'", id, batchID)
	return nil
}
