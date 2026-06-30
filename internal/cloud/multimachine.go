package cloud

import "context"

type MultiMachineAdapter struct {
	downloadFunc func(ctx context.Context, machineID, remotePath, localPath string) error
	uploadFunc   func(ctx context.Context, machineID, localPath, remotePath string) error
}

func NewMultiMachineAdapter(
	downloadFunc func(ctx context.Context, machineID, remotePath, localPath string) error,
	uploadFunc func(ctx context.Context, machineID, localPath, remotePath string) error,
) *MultiMachineAdapter {
	return &MultiMachineAdapter{
		downloadFunc: downloadFunc,
		uploadFunc:   uploadFunc,
	}
}

func (a *MultiMachineAdapter) DownloadFromMachine(ctx context.Context, machineID, remotePath, localPath string) error {
	if a.downloadFunc != nil {
		return a.downloadFunc(ctx, machineID, remotePath, localPath)
	}
	return nil
}

func (a *MultiMachineAdapter) UploadToMachine(ctx context.Context, machineID, localPath, remotePath string) error {
	if a.uploadFunc != nil {
		return a.uploadFunc(ctx, machineID, localPath, remotePath)
	}
	return nil
}
