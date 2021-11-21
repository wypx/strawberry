package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryMigrationExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryMigrationExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryMigrationResponse)
	resp.SetSuccess(false)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.QueryMigration(respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] query migration fail: %s", err.Error())
		resp.SetError(err.Error())
		return nil
	}
	var idList, errMessage []string
	var finish, progress []uint64
	for _, m := range result.MigrationList {
		idList = append(idList, m.ID)
		if m.Finished {
			finish = append(finish, 1)
		} else {
			finish = append(finish, 0)
		}
		progress = append(progress, uint64(m.Progress))
		if m.Error != nil {
			errMessage = append(errMessage, m.Error.Error())
		} else {
			errMessage = append(errMessage, "")
		}

	}
	resp.SetSuccess(true)
	resp.SetStringArray(vm_utils.ParamKeyMigration, idList)
	resp.SetUIntArray(vm_utils.ParamKeyStatus, finish)
	resp.SetUIntArray(vm_utils.ParamKeyProgress, progress)
	resp.SetStringArray(vm_utils.ParamKeyError, errMessage)
	log.Printf("[%08X] %d migrations available", id, len(idList))
	return executor.Sender.SendMessage(resp, request.GetSender())
}
