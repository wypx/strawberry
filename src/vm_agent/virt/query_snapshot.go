package virt

import (
	"errors"
	"log"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type QuerySnapshotExecutor struct {
	Sender        VmUtils.MessageSender
	StorageModule VmAgentSvc.StorageModule
}

func (executor *QuerySnapshotExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	if instanceID, err = request.GetString(VmUtils.ParamKeyInstance); err != nil {
		return err
	}

	log.Printf("[%08X] recv query snapshots for guest '%s' from %s.[%08X]",
		id, instanceID, request.GetSender(), request.GetFromSession())
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.QuerySnapshotResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	{
		var respChan = make(chan VmAgentSvc.StorageResult, 1)
		executor.StorageModule.QuerySnapshot(instanceID, respChan)
		var timer = time.NewTimer(VmAgentSvc.DefaultOperateTimeout)
		select {
		case <-timer.C:
			err = errors.New("request timeout")
			log.Printf("[%08X] query snapshot timeout", id)
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		case result := <-respChan:
			if result.Error != nil {
				err = result.Error
				log.Printf("[%08X] query snapshot fail: %s", id, err.Error())
				resp.SetError(err.Error())
			} else {
				var snapshotList = result.SnapshotList
				var names, backings []string
				var rootFlags, currentFlags []uint64
				for _, snapshot := range snapshotList {
					names = append(names, snapshot.Name)
					backings = append(backings, snapshot.Backing)
					if snapshot.IsRoot {
						rootFlags = append(rootFlags, 1)
					} else {
						rootFlags = append(rootFlags, 0)
					}
					if snapshot.IsCurrent {
						currentFlags = append(currentFlags, 1)
					} else {
						currentFlags = append(currentFlags, 0)
					}
				}
				resp.SetStringArray(VmUtils.ParamKeyName, names)
				resp.SetStringArray(VmUtils.ParamKeyPrevious, backings)
				resp.SetUIntArray(VmUtils.ParamKeySource, rootFlags)
				resp.SetUIntArray(VmUtils.ParamKeyCurrent, currentFlags)
				log.Printf("[%08X] %d snapshot(s) available for guest '%s'", id, len(snapshotList), instanceID)
				resp.SetSuccess(true)
			}
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
