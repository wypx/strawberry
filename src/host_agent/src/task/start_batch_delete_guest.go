package task

import (
	"errors"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type StartBatchDeleteGuestExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *StartBatchDeleteGuestExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var guestList []string
	if guestList, err = request.GetStringArray(vm_utils.ParamKeyGuest); err != nil {
		return err
	}
	log.Printf("[%08X] recv batch delete %d guests from %s.[%08X]", id, len(guestList), request.GetSender(), request.GetFromSession())

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.StartBatchDeleteGuestResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var batchID string
	{
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.StartBatchDeleteGuest(guestList, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] start batch delete guest fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		batchID = result.Batch
	}

	var targets = map[vm_utils.TransactionID]string{}
	for index, guestID := range guestList {
		var transID = vm_utils.TransactionID(index)
		deleteGuest, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteGuestRequest)
		deleteGuest.SetFromSession(id)
		deleteGuest.SetString(vm_utils.ParamKeyInstance, guestID)
		deleteGuest.SetTransactionID(transID)
		targets[transID] = guestID
		if err = executor.Sender.SendToSelf(deleteGuest); err != nil {
			log.Printf("[%08X] warning: request delete guest '%s' fail: %s", id, guestID, err.Error())
		}
		//log.Printf("[%08X] debug: trans %d => guest '%s'", id, transID, guestID)
	}
	log.Printf("[%08X] new batch delete '%s' started", id, batchID)
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
				log.Printf("[%08X] warning: receive delete response timeout", id)
				return
			}
		case deleteResponse := <-incoming:
			var transID = deleteResponse.GetTransactionID()
			guestID, exists := targets[transID]
			if !exists {
				log.Printf("[%08X] warning: invalid delete response with trans [%08X] from [%08X]",
					id, transID, deleteResponse.GetFromSession())
				break
			}
			var errChan = make(chan error, 1)
			if deleteResponse.IsSuccess() {
				log.Printf("[%08X] guest '%s' deleted", id, guestID)
				executor.ResourceModule.SetBatchDeleteGuestSuccess(batchID, guestID, errChan)
			} else {
				var deleteError = errors.New(deleteResponse.GetError())
				log.Printf("[%08X] delete guest '%s' fail: %s", id, guestID, deleteError.Error())
				executor.ResourceModule.SetBatchDeleteGuestFail(batchID, guestID, deleteError, errChan)
			}
			var result = <-errChan
			if result != nil {
				log.Printf("[%08X] warning:update delete status fail: %s", id, result.Error())
			}
			lastUpdate = time.Now()
			delete(targets, transID)
		}
	}
	//all targets finished
	log.Printf("[%08X] all delete request finished in batch '%s'", id, batchID)
	return nil
}
