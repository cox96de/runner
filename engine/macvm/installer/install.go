package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Code-Hex/vz/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func getInstallCommand() *cobra.Command {
	var (
		restoreImage string
		output       string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "install a macOS image to a disk image",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%+v\n", restoreImage)
			fmt.Printf("%+v\n", output)
			return install(context.Background(), restoreImage, output, 0, 0)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&restoreImage, "restore-image", "", "restore-image image ipsw file path")
	flags.StringVarP(&output, "output", "o", "VM.bundle", "output file path")
	return cmd
}

func install(ctx context.Context, restoreImagePath string, output string, cpu uint64, memory uint64) error {
	if _, err := os.Stat(restoreImagePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// TODO: download restart image
	}
	restoreImage, err := vz.LoadMacOSRestoreImageFromPath(restoreImagePath)
	if err != nil {
		return errors.WithStack(err)
	}
	err = os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return err
	}
	bootLoader, err := vz.NewMacOSBootLoader()
	if err != nil {
		return err
	}
	mostFeaturefulSupportedConfiguration := restoreImage.MostFeaturefulSupportedConfiguration()
	cpuCount := computeCPUCount()
	memorySize := computeMemorySize()
	machineConfiguration, err := vz.NewVirtualMachineConfiguration(bootLoader, cpuCount, memorySize)
	if err != nil {
		return err
	}
	platformConfiguration, err := createMacInstallerPlatformConfiguration(output, mostFeaturefulSupportedConfiguration)
	if err != nil {
		return err
	}
	machineConfiguration.SetPlatformVirtualMachineConfiguration(platformConfiguration)
	diskImagePath := GetDiskImagePath(output)
	err = vz.CreateDiskImage(diskImagePath, 32<<30)
	if err != nil {
		return err
	}
	storageDeviceAttachment, err := vz.NewDiskImageStorageDeviceAttachment(diskImagePath, false)
	if err != nil {
		return err
	}
	blockDeviceConfiguration, err := vz.NewVirtioBlockDeviceConfiguration(storageDeviceAttachment)
	if err != nil {
		return err
	}
	machineConfiguration.SetStorageDevicesVirtualMachineConfiguration([]vz.StorageDeviceConfiguration{blockDeviceConfiguration})
	vm, err := vz.NewVirtualMachine(machineConfiguration)
	if err != nil {
		return err
	}
	installer, err := vz.NewMacOSInstaller(vm, restoreImagePath)
	if err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("install has been cancelled")
				return
			case <-installer.Done():
				log.Println("install has been completed")
				return
			case <-ticker.C:
				log.Printf("install: %.3f%%\r", installer.FractionCompleted()*100)
			}
		}
	}()
	return installer.Install(ctx)
}

func computeCPUCount() uint {
	totalAvailableCPUs := runtime.NumCPU()
	virtualCPUCount := uint(totalAvailableCPUs - 1)
	if virtualCPUCount <= 1 {
		virtualCPUCount = 1
	}
	// TODO(codehex): use generics function when deprecated Go 1.17
	maxAllowed := vz.VirtualMachineConfigurationMaximumAllowedCPUCount()
	if virtualCPUCount > maxAllowed {
		virtualCPUCount = maxAllowed
	}
	minAllowed := vz.VirtualMachineConfigurationMinimumAllowedCPUCount()
	if virtualCPUCount < minAllowed {
		virtualCPUCount = minAllowed
	}
	return virtualCPUCount
}

func computeMemorySize() uint64 {
	// We arbitrarily choose 4GB.
	memorySize := uint64(4 * 1024 * 1024 * 1024)
	maxAllowed := vz.VirtualMachineConfigurationMaximumAllowedMemorySize()
	if memorySize > maxAllowed {
		memorySize = maxAllowed
	}
	minAllowed := vz.VirtualMachineConfigurationMinimumAllowedMemorySize()
	if memorySize < minAllowed {
		memorySize = minAllowed
	}
	return memorySize
}

func createMacInstallerPlatformConfiguration(output string,
	macOSConfiguration *vz.MacOSConfigurationRequirements,
) (*vz.MacPlatformConfiguration, error) {
	hardwareModel := macOSConfiguration.HardwareModel()
	if err := CreateFileAndWriteTo(
		hardwareModel.DataRepresentation(),
		GetHardwareModelPath(output),
	); err != nil {
		return nil, fmt.Errorf("failed to write hardware model data: %w", err)
	}

	machineIdentifier, err := vz.NewMacMachineIdentifier()
	if err != nil {
		return nil, err
	}
	if err := CreateFileAndWriteTo(
		machineIdentifier.DataRepresentation(),
		GetMachineIdentifierPath(output),
	); err != nil {
		return nil, fmt.Errorf("failed to write machine identifier data: %w", err)
	}

	auxiliaryStorage, err := vz.NewMacAuxiliaryStorage(
		GetAuxiliaryStoragePath(output),
		vz.WithCreatingMacAuxiliaryStorage(hardwareModel),
	)
	platformConfiguration, err := vz.NewMacPlatformConfiguration()
	platformConfiguration.HardwareModel()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new mac auxiliary storage: %w", err)
	}
	return vz.NewMacPlatformConfiguration(
		vz.WithMacAuxiliaryStorage(auxiliaryStorage),
		vz.WithMacHardwareModel(hardwareModel),
		vz.WithMacMachineIdentifier(machineIdentifier),
	)
}

// GetAuxiliaryStoragePath gets a path for auxiliary storage.
func GetAuxiliaryStoragePath(output string) string {
	return filepath.Join(output, "AuxiliaryStorage")
}

// GetDiskImagePath gets a path for disk image.
func GetDiskImagePath(output string) string {
	return filepath.Join(output, "Disk.img")
}

// GetHardwareModelPath gets a path for hardware model.
func GetHardwareModelPath(output string) string {
	return filepath.Join(output, "HardwareModel")
}

// GetMachineIdentifierPath gets a path for machine identifier.
func GetMachineIdentifierPath(output string) string {
	return filepath.Join(output, "MachineIdentifier")
}

// CreateFileAndWriteTo creates a new file and write data to it.
func CreateFileAndWriteTo(data []byte, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	return nil
}
