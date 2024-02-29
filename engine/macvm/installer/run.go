package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Code-Hex/vz/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func getRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run a macOS virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("target is required")
			}
			target := args[0]
			return runVM(context.Background(), target)
		},
	}
	return cmd
}

func handleConn(incomingConn net.Conn, socketDevice *vz.VirtioSocketDevice) {
	var (
		connect *vz.VirtioSocketConnection
		err     error
	)
	for i := 0; i < 100; i++ {
		connect, err = socketDevice.Connect(2222)
		if err != nil {
			nsError, ok := err.(*vz.NSError)
			if ok && nsError.Code == int(syscall.ECONNRESET) {
				log.Println("failed to connect with reset", err)
				time.Sleep(time.Second)
				continue
			}
			log.Println("failed to connect", err)
			return
		}
		break
	}
	if connect == nil {
		fmt.Printf("failed to connect\n")
		return
	}
	defer connect.Close()
	closeCh := make(chan struct{}, 1)
	go func() {
		io.Copy(connect, incomingConn)
		close(closeCh)
	}()
	io.Copy(incomingConn, connect)
	<-closeCh
	incomingConn.Close()
	connect.Close()
}

func runVM(ctx context.Context, dir string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	platformConfig, err := createMacPlatformConfiguration(dir)
	if err != nil {
		return err
	}
	config, err := setupVMConfiguration(platformConfig, dir)
	if err != nil {
		return err
	}
	vm, err := vz.NewVirtualMachine(config)
	if err != nil {
		return err
	}
	socketDevices := vm.SocketDevices()
	if len(socketDevices) == 0 {
		return errors.Errorf("not found socket device")
	}
	socketDevice := socketDevices[0]
	go func() {
		listen, err := net.Listen("tcp", ":2223")
		if err != nil {
			log.Println("failed to listen", err)
			time.Sleep(time.Second)
			return
		}
		for {
			incomingConn, err := listen.Accept()
			if err != nil {
				log.Println("failed to accept", err)
				time.Sleep(time.Second)
				return
			}
			go handleConn(incomingConn, socketDevice)

		}
	}()

	if err := vm.Start(); err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		for {
			select {
			case newState := <-vm.StateChangedNotify():
				if newState == vz.VirtualMachineStateRunning {
					log.Println("start VM is running")
				}
				if newState == vz.VirtualMachineStateStopped || newState == vz.VirtualMachineStateStopping {
					log.Println("stopped state")
					errCh <- nil
					return
				}
			case err := <-errCh:
				errCh <- fmt.Errorf("failed to start vm: %w", err)
				return
			}
		}
	}()

	// cleanup is this function is useful when finished graphic application.
	cleanup := func() {
		for i := 1; vm.CanRequestStop(); i++ {
			result, err := vm.RequestStop()
			log.Printf("sent stop request(%d): %t, %v", i, result, err)
			time.Sleep(time.Second * 3)
			if i > 3 {
				log.Println("call stop")
				if err := vm.Stop(); err != nil {
					log.Println("stop with error", err)
				}
			}
		}
		log.Println("finished cleanup")
	}

	runtime.LockOSThread()
	if err = vm.StartGraphicApplication(960, 600); err != nil {
		return errors.WithMessage(err, "failed to start graphic application")
	}
	runtime.UnlockOSThread()
	cleanup()

	return <-errCh
}

func createBlockDeviceConfiguration(diskPath string) (*vz.VirtioBlockDeviceConfiguration, error) {
	diskImageAttachment, err := vz.NewDiskImageStorageDeviceAttachmentWithCacheAndSync(
		diskPath,
		false, vz.DiskImageCachingModeCached, vz.DiskImageSynchronizationModeNone,
	)
	if err != nil {
		return nil, err
	}
	return vz.NewVirtioBlockDeviceConfiguration(diskImageAttachment)
}

func createGraphicsDeviceConfiguration() (*vz.MacGraphicsDeviceConfiguration, error) {
	graphicDeviceConfig, err := vz.NewMacGraphicsDeviceConfiguration()
	if err != nil {
		return nil, err
	}
	graphicsDisplayConfig, err := vz.NewMacGraphicsDisplayConfiguration(1920, 1200, 80)
	if err != nil {
		return nil, err
	}
	graphicDeviceConfig.SetDisplays(
		graphicsDisplayConfig,
	)
	return graphicDeviceConfig, nil
}

func createNetworkDeviceConfiguration() (*vz.VirtioNetworkDeviceConfiguration, error) {
	natAttachment, err := vz.NewNATNetworkDeviceAttachment()
	if err != nil {
		return nil, err
	}
	return vz.NewVirtioNetworkDeviceConfiguration(natAttachment)
}

func createKeyboardConfiguration() (*vz.USBKeyboardConfiguration, error) {
	return vz.NewUSBKeyboardConfiguration()
}

func createAudioDeviceConfiguration() (*vz.VirtioSoundDeviceConfiguration, error) {
	audioConfig, err := vz.NewVirtioSoundDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create sound device configuration: %w", err)
	}
	inputStream, err := vz.NewVirtioSoundDeviceHostInputStreamConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create input stream configuration: %w", err)
	}
	outputStream, err := vz.NewVirtioSoundDeviceHostOutputStreamConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create output stream configuration: %w", err)
	}
	audioConfig.SetStreams(
		inputStream,
		outputStream,
	)
	return audioConfig, nil
}

func createMacPlatformConfiguration(dir string) (*vz.MacPlatformConfiguration, error) {
	auxiliaryStorage, err := vz.NewMacAuxiliaryStorage(GetAuxiliaryStoragePath(dir))
	if err != nil {
		return nil, fmt.Errorf("failed to create a new mac auxiliary storage: %w", err)
	}
	hardwareModel, err := vz.NewMacHardwareModelWithDataPath(
		GetHardwareModelPath(dir),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new hardware model: %w", err)
	}
	//machineIdentifier, err := vz.NewMacMachineIdentifierWithDataPath(
	//	GetMachineIdentifierPath(dir),
	//)
	machineIdentifier, err := vz.NewMacMachineIdentifierWithDataPath(GetMachineIdentifierPath(dir))
	//machineIdentifier, err := vz.NewMacMachineIdentifier()
	//if err != nil {
	//	return nil, err
	//}
	if err != nil {
		return nil, fmt.Errorf("failed to create a new machine identifier: %w", err)
	}
	return vz.NewMacPlatformConfiguration(
		vz.WithMacAuxiliaryStorage(auxiliaryStorage),
		vz.WithMacHardwareModel(hardwareModel),
		vz.WithMacMachineIdentifier(machineIdentifier),
	)
}

func setupVMConfiguration(platformConfig vz.PlatformConfiguration, dir string) (*vz.VirtualMachineConfiguration, error) {
	bootloader, err := vz.NewMacOSBootLoader()
	if err != nil {
		return nil, err
	}

	config, err := vz.NewVirtualMachineConfiguration(
		bootloader,
		computeCPUCount(),
		computeMemorySize(),
	)
	if err != nil {
		return nil, err
	}
	config.SetPlatformVirtualMachineConfiguration(platformConfig)
	graphicsDeviceConfig, err := createGraphicsDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create graphics device configuration: %w", err)
	}
	config.SetGraphicsDevicesVirtualMachineConfiguration([]vz.GraphicsDeviceConfiguration{
		graphicsDeviceConfig,
	})
	blockDeviceConfig, err := createBlockDeviceConfiguration(GetDiskImagePath(dir))
	if err != nil {
		return nil, fmt.Errorf("failed to create block device configuration: %w", err)
	}
	config.SetStorageDevicesVirtualMachineConfiguration([]vz.StorageDeviceConfiguration{
		blockDeviceConfig,
	})

	// Setup socket device.
	socketDeviceConfiguration, err := vz.NewVirtioSocketDeviceConfiguration()
	if err != nil {
		return nil, err
	}
	config.SetSocketDevicesVirtualMachineConfiguration([]vz.SocketDeviceConfiguration{socketDeviceConfiguration})
	networkDeviceConfig, err := createNetworkDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create network device configuration: %w", err)
	}
	macAddress, err := vz.NewRandomLocallyAdministeredMACAddress()
	if err != nil {
		return nil, err
	}
	fmt.Printf("the mac addr: %s\n", macAddress.String())
	networkDeviceConfig.SetMACAddress(macAddress)
	config.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
		networkDeviceConfig,
	})

	usbScreenPointingDevice, err := vz.NewUSBScreenCoordinatePointingDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create pointing device configuration: %w", err)
	}
	pointingDevices := []vz.PointingDeviceConfiguration{usbScreenPointingDevice}

	trackpad, err := vz.NewMacTrackpadConfiguration()
	if err == nil {
		pointingDevices = append(pointingDevices, trackpad)
	}
	config.SetPointingDevicesVirtualMachineConfiguration(pointingDevices)

	keyboardDeviceConfig, err := createKeyboardConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create keyboard device configuration: %w", err)
	}
	config.SetKeyboardsVirtualMachineConfiguration([]vz.KeyboardConfiguration{
		keyboardDeviceConfig,
	})

	audioDeviceConfig, err := createAudioDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create audio device configuration: %w", err)
	}
	config.SetAudioDevicesVirtualMachineConfiguration([]vz.AudioDeviceConfiguration{
		audioDeviceConfig,
	})
	//shareDirConfiguration, err := createShareDirConfiguration()
	//if err != nil {
	//	return nil, err
	//}
	//config.SetDirectorySharingDevicesVirtualMachineConfiguration(
	//	[]vz.DirectorySharingDeviceConfiguration{shareDirConfiguration})
	validated, err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}
	if !validated {
		return nil, fmt.Errorf("invalid configuration")
	}

	return config, nil
}
