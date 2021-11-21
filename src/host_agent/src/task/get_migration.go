package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetMigrationExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetMigrationExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var migrationID string
	migrationID, err = request.GetString(vm_utils.ParamKeyMigration)
	if err != nil {
		return
	}
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetMigrationResponse)
	resp.SetSuccess(false)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)

	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetMigration(migrationID, respChan)
	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] get migration fail: %s", id, err.Error())
		resp.SetError(err.Error())
		return nil
	}
	var migration = result.Migration
	resp.SetSuccess(true)
	resp.SetString(vm_utils.ParamKeyMigration, migrationID)
	resp.SetBoolean(vm_utils.ParamKeyStatus, migration.Finished)
	resp.SetUInt(vm_utils.ParamKeyProgress, migration.Progress)
	if migration.Error != nil {
		resp.SetString(vm_utils.ParamKeyError, migration.Error.Error())
	} else {
		resp.SetString(vm_utils.ParamKeyError, "")
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
