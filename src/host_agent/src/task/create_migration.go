package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"

	"github.com/pkg/errors"
)

type CreateMigrationExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *CreateMigrationExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var poolName, sourceCell, targetCell string
	var instances []string
	{
		const (
			SourcePoolOffset = 0
			SourceCellOffset = 0
			TargetCellOffset = 1
			ValidPoolCount   = 1
			ValidCellCount   = 2
		)
		var pools, cells []string
		if pools, err = request.GetStringArray(vm_utils.ParamKeyPool); err != nil {
			return
		}
		if cells, err = request.GetStringArray(vm_utils.ParamKeyCell); err != nil {
			return
		}
		if instances, err = request.GetStringArray(vm_utils.ParamKeyInstance); err != nil {
			return
		}
		if ValidPoolCount != len(pools) {
			err = fmt.Errorf("invalid migration pool count %d", len(pools))
			return
		}
		if ValidCellCount != len(cells) {
			err = fmt.Errorf("invalid migration cell count %d", len(cells))
			return
		}
		poolName = pools[SourcePoolOffset]
		sourceCell = cells[SourceCellOffset]
		targetCell = cells[TargetCellOffset]
	}
	if 0 == len(instances) {
		log.Printf("[%08X] request migrate all instance in '%s.%s' to '%s.%s' from %s.[%08X]",
			id, poolName, sourceCell, poolName, targetCell, request.GetSender(), request.GetFromSession())
	} else {
		log.Printf("[%08X] request migrate %d instance(s) in '%s.%s' to '%s.%s' from %s.[%08X]",
			id, len(instances), poolName, sourceCell, poolName, targetCell, request.GetSender(), request.GetFromSession())
	}

	var migrationID string
	var params = modules.MigrationParameter{SourcePool: poolName, SourceCell: sourceCell, TargetPool: poolName, TargetCell: targetCell, Instances: instances}
	{
		resp, _ := vm_utils.CreateJsonMessage(vm_utils.CreateMigrationResponse)
		resp.SetSuccess(false)
		resp.SetFromSession(id)
		resp.SetToSession(request.GetFromSession())

		//allocate migration
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.CreateMigration(params, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			resp.SetError(err.Error())
			log.Printf("[%08X] allocate migration task fail: %s", id, err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		migrationID = result.Migration.ID
		instances = result.Migration.Instances
		resp.SetSuccess(true)
		resp.SetString(vm_utils.ParamKeyMigration, migrationID)
		log.Printf("[%08X] migration '%s' allocated", id, migrationID)
		if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
			log.Printf("[%08X] warning: notify migration id fail: %s", id, err.Error())
		}
	}
	var targetSession vm_utils.SessionID
	{
		//attach instance
		attach, _ := vm_utils.CreateJsonMessage(vm_utils.AttachInstanceRequest)
		attach.SetFromSession(id)
		attach.SetBoolean(vm_utils.ParamKeyImmediate, false)
		attach.SetStringArray(vm_utils.ParamKeyInstance, instances)
		if err = executor.Sender.SendMessage(attach, targetCell); err != nil {
			log.Printf("[%08X] request attach instance fail: %s", id, err.Error())
			executor.releaseMigration(id, migrationID, err)
			return nil
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if !cellResp.IsSuccess() {
				//attach fail
				log.Printf("[%08X] attach instances fail: %s", id, cellResp.GetError())
				executor.releaseMigration(id, migrationID, fmt.Errorf("attach instance fail: %s", cellResp.GetError()))
				return nil
			}
			log.Printf("[%08X] instances attached to '%s.%s'", id, poolName, targetCell)
			targetSession = cellResp.GetFromSession()
		case <-timer.C:
			//timeout
			log.Printf("[%08X] attach instances timeout", id)
			executor.releaseMigration(id, migrationID, errors.New("attach instance timeout"))
			return nil
		}
	}
	{
		//detach
		detach, _ := vm_utils.CreateJsonMessage(vm_utils.DetachInstanceRequest)
		detach.SetFromSession(id)
		detach.SetStringArray(vm_utils.ParamKeyInstance, instances)
		if err = executor.Sender.SendMessage(detach, sourceCell); err != nil {
			log.Printf("[%08X] request detach instance fail: %s", id, err.Error())
			executor.releaseMigration(id, migrationID, err)
			return nil
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if !cellResp.IsSuccess() {
				//detach fail
				log.Printf("[%08X] detach instances fail: %s", id, cellResp.GetError())
				executor.releaseMigration(id, migrationID, fmt.Errorf("detach instance fail: %s", cellResp.GetError()))
				return nil
			}
			log.Printf("[%08X] instances detached from '%s.%s'", id, poolName, sourceCell)
		case <-timer.C:
			//timeout
			log.Printf("[%08X] detach instances timeout", id)
			executor.releaseMigration(id, migrationID, errors.New("detach instance timeout"))
			return nil
		}
	}
	{
		//migrate
		migrate, _ := vm_utils.CreateJsonMessage(vm_utils.MigrateInstanceRequest)
		migrate.SetFromSession(id)
		migrate.SetToSession(targetSession)
		migrate.SetString(vm_utils.ParamKeyMigration, migrationID)
		migrate.SetStringArray(vm_utils.ParamKeyInstance, instances)
		if err = executor.Sender.SendMessage(migrate, targetCell); err != nil {
			log.Printf("[%08X] warning: notify migrate fail: %s", id, err.Error())
			executor.releaseMigration(id, migrationID, err)
		} else {
			log.Printf("[%08X] notify '%s.%s' start migrate", id, poolName, targetCell)
		}
	}
	return nil
}

func (executor *CreateMigrationExecutor) releaseMigration(id vm_utils.SessionID, migration string, reason error) {
	var respChan = make(chan error, 1)
	executor.ResourceModule.CancelMigration(migration, reason, respChan)
	var err = <-respChan
	if err != nil {
		log.Printf("[%08X] warning: release migration fail: %s", id, migration)
	} else {
		log.Printf("[%08X] migration %s released", id, migration)
	}
}
