package imageserver

import (
	"fmt"
	"vm_manager/vm_utils"
)

type TaskManager struct {
	*vm_utils.TransactionEngine
}

func CreateTaskManager(sender vm_utils.MessageSender, imageManager *ImageManager) (*TaskManager, error) {
	engine, err := vm_utils.CreateTransactionEngine()
	if err != nil {
		return nil, err
	}

	var manager = TaskManager{engine}

	if err = manager.RegisterExecutor(vm_utils.QueryMediaImageRequest,
		&QueryMediaImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetMediaImageRequest,
		&GetMediaImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateMediaImageRequest,
		&CreateMediaImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyMediaImageRequest,
		&ModifyMediaImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteMediaImageRequest,
		&DeleteMediaImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}

	if err = manager.RegisterExecutor(vm_utils.QueryDiskImageRequest,
		&QueryDiskImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.GetDiskImageRequest,
		&GetDiskImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.CreateDiskImageRequest,
		&CreateDiskImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.ModifyDiskImageRequest,
		&ModifyDiskImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DeleteDiskImageRequest,
		&DeleteDiskImageExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.DiskImageUpdatedEvent,
		&DiskImageUpdateExecutor{sender, imageManager}); err != nil {
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.SynchronizeDiskImageRequest,
		&SyncDiskImagesExecutor{
			Sender:      sender,
			ImageServer: imageManager,
		}); err != nil {
		err = fmt.Errorf("register sync disk images fail: %s", err.Error())
		return nil, err
	}
	if err = manager.RegisterExecutor(vm_utils.SynchronizeMediaImageRequest,
		&SyncMediaImagesExecutor{
			Sender:      sender,
			ImageServer: imageManager,
		}); err != nil {
		err = fmt.Errorf("register sync disk images fail: %s", err.Error())
		return nil, err
	}

	return &manager, nil
}
