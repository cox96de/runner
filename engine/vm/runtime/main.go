//go:build linux

package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cox96de/containervm/network"
	"github.com/cox96de/containervm/util"
	util2 "github.com/cox96de/runner/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type config struct {
	ExecutorPath         string
	CloudInitMetaDataEnv string
	CloudInitUserDataEnv string
	Args                 []string
}

func parseConfig(args []string) (*config, error) {
	c := &config{}
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	flagSet.StringVar(&c.ExecutorPath, "executor-path", "/executor", "The path of the executor")
	flagSet.StringVar(&c.CloudInitUserDataEnv, "cloud-init-user-data-env", "", "The env var name of cloud init user data")
	flagSet.StringVar(&c.CloudInitMetaDataEnv, "cloud-init-meta-data-env", "", "The env var name of cloud init meta data")
	err := flagSet.Parse(args)
	if err != nil {
		return nil, err
	}
	c.Args = flagSet.Args()
	return c, nil
}

func generateCloudInitISO(metaData, userData string) (string, error) {
	tempDir, err := os.MkdirTemp("", "cloud-init")
	if err != nil {
		return "", errors.WithMessage(err, "failed to create temp dir")
	}
	metaDataPath := filepath.Join(tempDir, "meta-data")
	err = os.WriteFile(metaDataPath, []byte(metaData), os.ModePerm)
	if err != nil {
		return "", errors.WithMessagef(err, "failed to write meta data to '%s'", metaDataPath)
	}
	userDataPath := filepath.Join(tempDir, "user-data")
	err = os.WriteFile(userDataPath, []byte(userData), os.ModePerm)
	if err != nil {
		return "", errors.WithMessagef(err, "failed to write user data to '%s'", userDataPath)
	}
	isoFilepath := filepath.Join(tempDir, "cloud-init.iso")
	err = genISO("cidata", isoFilepath, metaDataPath, userDataPath)
	if err != nil {
		return "", errors.WithMessagef(err, "failed to generate cloud init iso")
	}
	return isoFilepath, nil
}

func main() {
	log.SetLevel(log.DebugLevel)
	c, err := parseConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse config: %+v", err)
	}
	args := c.Args
	fmt.Printf("%+v\n", args)
	fmt.Printf("%+v\n", os.Args[1:])
	var cloudInitISOPath string
	if len(c.CloudInitUserDataEnv) > 0 || len(c.CloudInitMetaDataEnv) > 0 {
		log.Infof("generating cloud init iso file")
		metadata := os.Getenv(c.CloudInitMetaDataEnv)
		userdata := os.Getenv(c.CloudInitUserDataEnv)
		log.Debugf("metadata: %s", metadata)
		log.Debugf("userdata: %s", userdata)
		cloudInitISOPath, err = generateCloudInitISO(metadata, userdata)
		if err != nil {
			log.Fatalf("failed to generate cloud init iso file: %+v", err)
		}
	}
	tapDevicePath, bridgeMacAddr, mtu := configureNetwork()
	tapFile, err := os.Open(tapDevicePath)
	if err != nil {
		log.Fatalf("failed to open tap dev(%s): %+v", tapDevicePath, err)
	}
	qemuNetworkOpt := generateQEMUNetworkOpt(tapFile, bridgeMacAddr, mtu)
	args = append(args, qemuNetworkOpt...)
	if len(cloudInitISOPath) > 0 {
		args = append(args, generateCloudInitDeviceOpt(cloudInitISOPath)...)
	}
	log.Infof("run qemu with command: %s", strings.Join(args, " "))
	qemuCMD := exec.Command(args[0], args[1:]...)
	qemuCMD.Stdin = os.Stdin
	qemuCMD.Stdout = os.Stdout
	qemuCMD.Stderr = os.Stderr
	qemuCMD.ExtraFiles = []*os.File{tapFile}
	if err := qemuCMD.Start(); err != nil {
		log.Fatalf("failed to start qemu: %+v", err)
	}
	if err := qemuCMD.Wait(); err != nil {
		log.Fatalf("failed to wait for qemu: %+v", err)
	}
	log.Infof("qemu exited with code %d", qemuCMD.ProcessState.ExitCode())
}

func generateQEMUNetworkOpt(vtapFile *os.File, macAddr net.HardwareAddr, mtu int) []string {
	return []string{
		"-netdev", fmt.Sprintf("tap,id=net0,vhost=on,fd=%d", vtapFile.Fd()),
		"-device", "virtio-net-pci,netdev=net0,mac=" + macAddr.String() + ",host_mtu=" + strconv.Itoa(mtu),
	}
}

func generateCloudInitDeviceOpt(path string) []string {
	return []string{"-drive", fmt.Sprintf("file=%s,media=cdrom,format=raw,readonly=on,if=ide,aio=threads", path)}
}

func configureNetwork() (bridgeName string, bridgeMacAddr net.HardwareAddr, mtu int) {
	nic, err := util.GetDefaultNIC()
	if err != nil {
		log.Fatalf("failed to get default nic: %+v", err)
	}
	log.Infof("reconfiguring nic %s", nic.Name)
	tapName := fmt.Sprintf("macvtap%s", util2.RandomString(3))
	lanName := fmt.Sprintf("macvlan%s", util2.RandomString(3))
	tapDevicePath := "/dev/" + tapName
	err = network.CreateBridge(&network.CreateBridgeOption{
		NICName:       nic.Name,
		NICMac:        nic.HardwareAddr,
		NewNICMac:     util.GetRandomMAC(),
		TapName:       tapName,
		TapDevicePath: tapDevicePath,
		LanName:       lanName,
	})
	if err != nil {
		log.Fatalf("failed to set up bridge: %+v", err)
	}
	log.Infof("tap device %s is created", tapName)
	// Start a DHCP server.
	hostname, _ := os.Hostname()
	ds, err := network.NewDHCPServerFromAddr(&network.DHCPOption{
		HardwareAddr:  nic.HardwareAddr,
		IP:            nic.Addr,
		GatewayIP:     nic.Gateway,
		DNSServers:    []string{},
		SearchDomains: []string{},
		Hostname:      hostname,
	})
	if err != nil {
		log.Fatalf("failed to create dhcp server: %+v", err)
	}
	go func() {
		if err := ds.Run(lanName); err != nil {
			log.Errorf("failed to start dhcp server: %+v", err)
		}
	}()
	go func() {
		if err := network.ServeARP(lanName, nic.Addr, nic.HardwareAddr, nic.GatewayHardwareAddr); err != nil {
			log.Errorf("failed to start arp server: %+v", err)
		}
	}()
	return tapDevicePath, nic.HardwareAddr, nic.MTU
}
