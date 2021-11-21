package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleInstanceMigratedExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *HandleInstanceMigratedExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var failover = false
	var migrationID string
	var instances []string
	var monitorPorts []uint64
	failover, err = request.GetBoolean(vm_utils.ParamKeyImmediate)
	if err != nil {
		return
	}
	if instances, err = request.GetStringArray(vm_utils.ParamKeyInstance); err != nil {
		return
	}
	if monitorPorts, err = request.GetUIntArray(vm_utils.ParamKeyMonitor); err != nil {
		return
	}
	if !failover {
		//active migration
		migrationID, err = request.GetString(vm_utils.ParamKeyMigration)
		if err != nil {
			return
		}

		var respChan = make(chan error, 1)
		executor.ResourceModule.FinishMigration(migrationID, instances, monitorPorts, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] finish migration fail: %s", id, err.Error())
		} else {
			log.Printf("[%08X] migration '%s' finished from %s.[%08X]", id, migrationID, request.GetSender(), request.GetFromSession())
		}
		return nil
	} else {
		//failover
		sourceCell, err := request.GetString(vm_utils.ParamKeyCell)
		if err != nil {
			return err
		}
		var respChan = make(chan error, 1)
		executor.ResourceModule.MigrateInstance(sourceCell, request.GetSender(), instances, monitorPorts, respChan)
		err = <-respChan
		if err != nil {
			log.Printf("[%08X] migrate instance fail: %s", id, err.Error())
		} else {
			log.Printf("[%08X] %d instance(s) migrated from '%s' to '%s'", id, len(instances), sourceCell, request.GetSender())
		}
		return nil
	}

}
