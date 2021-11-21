package task

import (
	"errors"
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type StartBatchStopGuestExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *StartBatchStopGuestExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var guestList []string
	if guestList, err = request.GetStringArray(vm_utils.ParamKeyGuest); err != nil {
		return err
	}
	options, err := request.GetUIntArray(vm_utils.ParamKeyOption)
	if err != nil {
		return err
	}
	const (
		ValidOptionCount = 2
	)
	if len(options) != ValidOptionCount {
		return fmt.Errorf("unexpected option count %d / %d", len(options), ValidOptionCount)
	}

	log.Printf("[%08X] recv batch stop %d guests from %s.[%08X]", id, len(guestList), request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.StartBatchStopGuestResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var batchID string
	{
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.StartBatchStopGuest(guestList, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] start batch stop guest fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		batchID = result.Batch
	}

	var targets = map[vm_utils.TransactionID]string{}
	for index, guestID := range guestList {
		var transID = vm_utils.TransactionID(index)
		stopGuest, _ := vm_utils.CreateJsonMessage(vm_utils.StopInstanceRequest)
		stopGuest.SetFromSession(id)
		stopGuest.SetString(vm_utils.ParamKeyInstance, guestID)
		stopGuest.SetUIntArray(vm_utils.ParamKeyOption, options)
		stopGuest.SetTransactionID(transID)
		targets[transID] = guestID
		if err = executor.Sender.SendToSelf(stopGuest); err != nil {
			log.Printf("[%08X] warning: request stop guest '%s' fail: %s", id, guestID, err.Error())
		}
	}
	log.Printf("[%08X] new batch stop '%s' started", id, batchID)
	resp.SetSuccess(true)
	resp.SetString(vm_utils.ParamKeyID, batchID)
	executor.Sender.SendMessage(resp, request.GetSender())

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
				log.Printf("[%08X] warning: receive stop response timeout", id)
				return
			}
		case stopResponse := <-incoming:
			var transID = stopResponse.GetTransactionID()
			guestID, exists := targets[transID]
			if !exists {
				log.Printf("[%08X] warning: invalid stop response with trans [%08X] from [%08X]",
					id, transID, stopResponse.GetFromSession())
				break
			}
			var errChan = make(chan error, 1)
			if stopResponse.IsSuccess() {
				log.Printf("[%08X] guest '%s' stopped", id, guestID)
				executor.ResourceModule.SetBatchStopGuestSuccess(batchID, guestID, errChan)
			} else {
				var stopError = errors.New(stopResponse.GetError())
				log.Printf("[%08X] stop guest '%s' fail: %s", id, guestID, stopError.Error())
				executor.ResourceModule.SetBatchStopGuestFail(batchID, guestID, stopError, errChan)
			}
			var result = <-errChan
			if result != nil {
				log.Printf("[%08X] warning:update stop status fail: %s", id, result.Error())
			}
			lastUpdate = time.Now()
			delete(targets, transID)
		}
	}
	//all targets finished
	log.Printf("[%08X] all stop request finished in batch '%s'", id, batchID)
	return nil
}
