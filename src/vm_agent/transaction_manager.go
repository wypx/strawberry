package vm_agent

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmAgentVirt "vm_manager/vm_agent/virt"
	VmUtils "vm_manager/vm_utils"
)

type TransactionManager struct {
	*VmUtils.TransactionEngine
}

func CreateTransactionManager(sender VmUtils.MessageSender, instanceModule *VmAgentSvc.InstanceManager,
	storageModule *VmAgentSvc.StorageManager, networkModule *VmAgentSvc.NetworkManager) (manager *TransactionManager, err error) {
	var engine *VmUtils.TransactionEngine
	if engine, err = VmUtils.CreateTransactionEngine(); err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	manager = &TransactionManager{engine}
	if err = manager.RegisterExecutor(VmUtils.GetComputePoolCellRequest,
		&VmAgentVirt.GetCellInfoExecutor{sender, instanceModule, storageModule, networkModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(VmUtils.CreateGuestRequest,
		&VmAgentVirt.CreateInstanceExecutor{sender, instanceModule, storageModule, networkModule, generator}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.DeleteGuestRequest,
		&VmAgentVirt.DeleteInstanceExecutor{sender, instanceModule, storageModule, networkModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.GetGuestRequest,
		&VmAgentVirt.GetInstanceConfigExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.GetInstanceStatusRequest,
		&VmAgentVirt.GetInstanceStatusExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.StartInstanceRequest,
		&VmAgentVirt.StartInstanceExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.StopInstanceRequest,
		&VmAgentVirt.StopInstanceExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.AttachInstanceRequest,
		&VmAgentVirt.AttachInstanceExecutor{sender, instanceModule, storageModule, networkModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.DetachInstanceRequest,
		&VmAgentVirt.DetachInstanceExecutor{sender, instanceModule, storageModule, networkModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyGuestNameRequest,
		&VmAgentVirt.ModifyGuestNameExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyCoreRequest,
		&VmAgentVirt.ModifyGuestCoreExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyMemoryRequest,
		&VmAgentVirt.ModifyGuestMemoryExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(VmUtils.ModifyPriorityRequest,
		&VmAgentVirt.ModifyCPUPriorityExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyDiskThresholdRequest,
		&VmAgentVirt.ModifyDiskThresholdExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyNetworkThresholdRequest,
		&VmAgentVirt.ModifyNetworkThresholdExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(VmUtils.ModifyAuthRequest,
		&VmAgentVirt.ModifyGuestPasswordExecutor{sender, instanceModule, generator}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.GetAuthRequest,
		&VmAgentVirt.GetGuestPasswordExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ResetSystemRequest,
		&VmAgentVirt.ResetGuestSystemExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.InsertMediaRequest,
		&VmAgentVirt.InsertMediaCoreExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.EjectMediaRequest,
		&VmAgentVirt.EjectMediaCoreExecutor{sender, instanceModule}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(VmUtils.ComputePoolReadyEvent,
		&VmAgentVirt.HandleComputePoolReadyExecutor{sender, instanceModule, storageModule, networkModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ComputeCellRemovedEvent,
		&VmAgentVirt.HandleComputeCellRemovedExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.CreateDiskImageRequest,
		&VmAgentVirt.CreateDiskImageExecutor{sender, instanceModule, storageModule, client}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ResizeDiskRequest,
		&VmAgentVirt.ResizeGuestVolumeExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ShrinkDiskRequest,
		&VmAgentVirt.ShrinkGuestVolumeExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.QuerySnapshotRequest,
		&VmAgentVirt.QuerySnapshotExecutor{sender, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.GetSnapshotRequest,
		&VmAgentVirt.GetSnapshotExecutor{sender, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.AddressPoolChangedEvent,
		&VmAgentVirt.HandleAddressPoolChangedExecutor{instanceModule, networkModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.CreateSnapshotRequest,
		&VmAgentVirt.CreateSnapshotExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.DeleteSnapshotRequest,
		&VmAgentVirt.DeleteSnapshotExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.RestoreSnapshotRequest,
		&VmAgentVirt.RestoreSnapshotExecutor{sender, instanceModule, storageModule}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(VmUtils.ResetSecretRequest,
		&VmAgentVirt.ResetMonitorSecretExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register reset monitor secret fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.QueryCellStorageRequest,
		&VmAgentVirt.QueryStoragePathExecutor{
			Sender:  sender,
			Storage: storageModule,
		}); err != nil {
		err = fmt.Errorf("register query storage paths fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyCellStorageRequest,
		&VmAgentVirt.ChangeStoragePathExecutor{
			Sender:  sender,
			Storage: storageModule,
		}); err != nil {
		err = fmt.Errorf("register change storage path fail: %s", err.Error())
		return
	}
	//security policy
	if err = manager.RegisterExecutor(VmUtils.GetGuestRuleRequest,
		&VmAgentVirt.GetSecurityPolicyExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register get security policy fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.AddGuestRuleRequest,
		&VmAgentVirt.AddSecurityRuleExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register add security rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyGuestRuleRequest,
		&VmAgentVirt.ModifySecurityRuleExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register modify security rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.ChangeGuestRuleDefaultActionRequest,
		&VmAgentVirt.ChangeDefaultSecurityActionExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register change default security action fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.ChangeGuestRuleOrderRequest,
		&VmAgentVirt.ChangeSecurityRuleOrderExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register change security rule order fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.RemoveGuestRuleRequest,
		&VmAgentVirt.RemoveSecurityRuleExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register remove security rule fail: %s", err.Error())
		return
	}
	if err = manager.RegisterExecutor(VmUtils.ModifyAutoStartRequest,
		&VmAgentVirt.ModifyAutoStartExecutor{
			Sender:         sender,
			InstanceModule: instanceModule,
		}); err != nil {
		err = fmt.Errorf("register modify auto start fail: %s", err.Error())
		return
	}
	return manager, nil
}
