package host_agent

import (
	"fmt"
	"vm_manager/host_agent/src/modules"
	"vm_manager/host_agent/src/task"
	"vm_manager/vm_utils"

	"crypto/tls"
	"net/http"
)

type CoreTransactionManager struct {
	*vm_utils.TransactionEngine
}

func CreateTransactionManager(sender vm_utils.MessageSender, resourceModule modules.ResourceModule) (manager *CoreTransactionManager, err error) {
	var engine *vm_utils.TransactionEngine
	if engine, err = vm_utils.CreateTransactionEngine(); err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	manager = &CoreTransactionManager{engine}
	if err = manager.RegisterExecutor(vm_utils.QueryComputePoolRequest,
		&task.QueryComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetComputePoolRequest,
		&task.GetComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.CreateComputePoolRequest,
		&task.CreateComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyComputePoolRequest,
		&task.ModifyComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteComputePoolRequest,
		&task.DeleteComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	//storage pools
	if err = manager.RegisterExecutor(vm_utils.QueryStoragePoolRequest,
		&task.QueryStoragePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetStoragePoolRequest,
		&task.GetStoragePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.CreateStoragePoolRequest,
		&task.CreateStoragePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyStoragePoolRequest,
		&task.ModifyStoragePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteStoragePoolRequest,
		&task.DeleteStoragePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	//address pool&range
	if err = manager.RegisterExecutor(vm_utils.QueryAddressPoolRequest,
		&task.QueryAddressPoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetAddressPoolRequest,
		&task.GetAddressPoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateAddressPoolRequest,
		&task.CreateAddressPoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyAddressPoolRequest,
		&task.ModifyAddressPoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteAddressPoolRequest,
		&task.DeleteAddressPoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryAddressRangeRequest,
		&task.QueryAddressRangeExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetAddressRangeRequest,
		&task.GetAddressRangeExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.AddAddressRangeRequest,
		&task.AddAddressRangeExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.RemoveAddressRangeRequest,
		&task.RemoveAddressRangeExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.QueryComputePoolCellRequest,
		&task.QueryCellsByPoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetComputePoolCellRequest,
		&task.GetComputeCellExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.AddComputePoolCellRequest,
		&task.AddComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.RemoveComputePoolCellRequest,
		&task.RemoveComputePoolExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryUnallocatedComputePoolCellRequest,
		&task.QueryUnallocatedCellsExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryZoneStatusRequest,
		&task.QueryZoneStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryComputePoolStatusRequest,
		&task.QueryComputePoolStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetComputePoolStatusRequest,
		&task.GetComputePoolStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryComputePoolCellStatusRequest,
		&task.QueryComputeCellStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetComputePoolCellStatusRequest,
		&task.GetComputeCellStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.EnableComputePoolCellRequest,
		&task.EnableComputeCellExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DisableComputePoolCellRequest,
		&task.DisableComputeCellExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ComputeCellAvailableEvent,
		&task.HandleCellAvailableExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetGuestRequest,
		&task.GetGuestConfigExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryGuestRequest,
		&task.QueryGuestConfigExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateGuestRequest,
		&task.CreateGuestExecutor{sender, resourceModule, client}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteGuestRequest,
		&task.DeleteGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyGuestNameRequest,
		&task.ModifyGuestNameExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyCoreRequest,
		&task.ModifyGuestCoreExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyMemoryRequest,
		&task.ModifyGuestMemoryExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.ModifyPriorityRequest,
		&task.ModifyGuestPriorityExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyNetworkThresholdRequest,
		&task.ModifyGuestNetworkThresholdExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyDiskThresholdRequest,
		&task.ModifyGuestDiskThresholdExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ResizeDiskRequest,
		&task.ResizeGuestDiskExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ShrinkDiskRequest,
		&task.ShrinkGuestDiskExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ResetSystemRequest,
		&task.ResetGuestSystemExecutor{sender, resourceModule, client}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyAuthRequest,
		&task.ModifyGuestPasswordExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetAuthRequest,
		&task.GetGuestPasswordExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.GetInstanceStatusRequest,
		&task.GetInstanceStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryInstanceStatusRequest,
		&task.QueryInstanceStatusExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.StartInstanceRequest,
		&task.StartInstanceExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.StopInstanceRequest,
		&task.StopInstanceExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	//media images
	if err = manager.RegisterExecutor(vm_utils.QueryMediaImageRequest,
		&task.QueryMediaImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetMediaImageRequest,
		&task.GetMediaImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateMediaImageRequest,
		&task.CreateMediaImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyMediaImageRequest,
		&task.ModifyMediaImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteMediaImageRequest,
		&task.DeleteMediaImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	//disk images
	if err = manager.RegisterExecutor(vm_utils.QueryDiskImageRequest,
		&task.QueryDiskImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetDiskImageRequest,
		&task.GetDiskImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateDiskImageRequest,
		&task.CreateDiskImageExecutor{sender, resourceModule, client}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyDiskImageRequest,
		&task.ModifyDiskImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteDiskImageRequest,
		&task.DeleteDiskImageExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.GuestCreatedEvent,
		&task.HandleGuestCreatedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GuestDeletedEvent,
		&task.HandleGuestDeletedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.GuestStartedEvent,
		&task.HandleGuestStartedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GuestStoppedEvent,
		&task.HandleGuestStoppedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GuestUpdatedEvent,
		&task.HandleGuestUpdatedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.SystemResetEvent,
		&task.HandleGuestSystemResetExecutor{resourceModule}); err != nil {
		return nil, err
	}
	//batch
	if err = manager.RegisterExecutor(vm_utils.StartBatchCreateGuestRequest,
		&task.StartBatchCreateGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetBatchCreateGuestRequest,
		&task.GetBatchCreateGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.StartBatchDeleteGuestRequest,
		&task.StartBatchDeleteGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetBatchDeleteGuestRequest,
		&task.GetBatchDeleteGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.StartBatchStopGuestRequest,
		&task.StartBatchStopGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetBatchStopGuestRequest,
		&task.GetBatchStopGuestExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	//instance media
	if err = manager.RegisterExecutor(vm_utils.InsertMediaRequest,
		&task.InsertMediaExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.EjectMediaRequest,
		&task.EjectMediaExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.MediaAttachedEvent,
		&task.HandleMediaAttachedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.MediaDetachedEvent,
		&task.HandleMediaDetachedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}

	//snapshot
	if err = manager.RegisterExecutor(vm_utils.QuerySnapshotRequest,
		&task.QuerySnapshotExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetSnapshotRequest,
		&task.GetSnapshotExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateSnapshotRequest,
		&task.CreateSnapshotExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteSnapshotRequest,
		&task.DeleteSnapshotExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.RestoreSnapshotRequest,
		&task.RestoreSnapshotExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.QueryMigrationRequest,
		&task.QueryMigrationExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetMigrationRequest,
		&task.GetMigrationExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateMigrationRequest,
		&task.CreateMigrationExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.InstanceMigratedEvent,
		&task.HandleInstanceMigratedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CellStatusReportEvent,
		&task.HandleCellStatusUpdatedExecutor{resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.AddressChangedEvent,
		&task.HandleAddressChangedExecutor{resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ComputeCellDisconnectedEvent,
		&task.HandleCellDisconnectedExecutor{sender, resourceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ImageServerAvailableEvent,
		&task.SyncImageServerExecutor{sender, resourceModule, client}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ResetSecretRequest,
		&task.ResetMonitorSecretExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register reset monitor secret fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.QueryCellStorageRequest,
		&task.QueryStoragePathsExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register query storage paths fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyCellStorageRequest,
		&task.ChangeStoragePathsExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register change storage path fail: %s", err.Error())
		return
	}
	//system templates
	if err = manager.RegisterExecutor(vm_utils.QueryTemplateRequest,
		&task.QuerySystemTemplatesExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register query system templates fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.GetTemplateRequest,
		&task.GetSystemTemplateExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register get system template fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.CreateTemplateRequest,
		&task.CreateSystemTemplateExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register create system template fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyTemplateRequest,
		&task.ModifySystemTemplateExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register modify system template fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteTemplateRequest,
		&task.DeleteSystemTemplateExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register delete system template fail: %s", err.Error())
		return
	}

	//Guest Security Policy Group
	if err = manager.RegisterExecutor(vm_utils.GetGuestRuleRequest,
		&task.GetGuestSecurityPolicyExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register get guest security policy fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ChangeGuestRuleDefaultActionRequest,
		&task.ChangeGuestSecurityActionExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register change guest security action fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ChangeGuestRuleOrderRequest,
		&task.ModifyGuestSecurityRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register move guest security rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.AddGuestRuleRequest,
		&task.AddGuestSecurityRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register add guest security rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyGuestRuleRequest,
		&task.ModifyGuestSecurityRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register modify guest security rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.RemoveGuestRuleRequest,
		&task.RemoveGuestSecurityRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register remove guest security rule fail: %s", err.Error())
		return
	}
	//Security Policy Group
	if err = manager.RegisterExecutor(vm_utils.QueryPolicyRuleRequest,
		&task.GetSecurityPolicyRulesExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register query security policy rules fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.AddPolicyRuleRequest,
		&task.AddSecurityPolicyRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register add security policy rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyPolicyRuleRequest,
		&task.ModifySecurityPolicyRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register modify security policy rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.RemovePolicyRuleRequest,
		&task.RemoveSecurityPolicyRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register remove security policy rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ChangePolicyRuleOrderRequest,
		&task.MoveSecurityPolicyRuleExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register move security policy rule fail: %s", err.Error())
		return
	}

	if err = manager.RegisterExecutor(vm_utils.QueryPolicyGroupRequest,
		&task.QuerySecurityPolicyGroupsExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register query security policy groups fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.GetPolicyGroupRequest,
		&task.GetSecurityPolicyGroupExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register get security policy group fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.CreatePolicyGroupRequest,
		&task.CreateSecurityPolicyGroupExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register create security policy group fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyPolicyGroupRequest,
		&task.ModifySecurityPolicyGroupExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register modify security policy group fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.DeletePolicyGroupRequest,
		&task.DeleteSecurityPolicyGroupExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register delete security policy group fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.SynchronizeMediaImageRequest,
		&task.SyncMediaImagesExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register sync media images fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.SynchronizeDiskImageRequest,
		&task.SyncDiskImagesExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register sync disk images fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.SearchGuestRequest,
		&task.SearchGuestsExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register search guests fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyAutoStartRequest,
		&task.ModifyGuestAutoStartExecutor{
			Sender:         sender,
			ResourceModule: resourceModule,
		}); err != nil {
		err = fmt.Errorf("register modify auto start fail: %s", err.Error())
		return
	}
	return manager, nil
}
