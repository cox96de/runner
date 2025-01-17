package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cox96de/containervm/network"
	"github.com/cox96de/runner/util"
	"github.com/spf13/cobra"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/containervm/cloudinit"
	vmutil "github.com/cox96de/containervm/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	defaultCloudInitMetadata = `#cloud-config
instance-id: someid/somehost
`
	defaultCloudInitUserData = `#cloud-config
`
)

const (
	userDataEnv = "CLOUD_INIT_USER_DATA"
	metaDataEnv = "CLOUD_INIT_META_DATA"
)

func main() {
	var consolePath string
	root := cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			qemu(cmd.Flags(), consolePath)
		},
	}
	pflag := root.Flags()
	pflag.StringVarP(&consolePath, "console", "c", "", "console socket path")
	root.AddCommand(&cobra.Command{
		Use: "socat",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				log.Fatalf("socat requires 2 arguments")
			}
			socat(args[0], args[1])
		},
	})
	checkError(root.Execute())
}

func qemu(pflag *pflag.FlagSet, consolePath string) {
	defaultNIC, err := vmutil.GetDefaultNIC()
	checkError(err, "failed to get default network interface")
	newMac := vmutil.GetRandomMAC()
	tapName := fmt.Sprintf("macvtap%s", util.RandomString(3))
	lanName := fmt.Sprintf("macvlan%s", util.RandomString(3))
	cloudinitConfig, err := generateCloudinitConfig(defaultNIC)
	checkError(err, "")
	bridgeConfigure := network.NewBridgeConfigure(defaultNIC.Name, newMac, tapName, lanName)
	err = bridgeConfigure.SetupBridge()
	checkError(err, "failed to setup bridge")
	defer func() {
		_ = bridgeConfigure.Recover()
	}()
	tapFile, err := os.Open(bridgeConfigure.GetMacVtapDevicePath())
	checkError(err, "failed to open tap device")
	log.Infof("network device %s", tapFile.Name())
	// FD is always 3, because 0, 1, 2 are reserved for stdin, stdout, stderr, the next available fd is 3.
	qemuNetworkOpt := generateQEMUNetworkOpt(3, defaultNIC.HardwareAddr, defaultNIC.MTU)
	var (
		cloudInitUserData = defaultCloudInitUserData
		cloudInitMetaData = defaultCloudInitMetadata
	)
	if e := os.Getenv(userDataEnv); len(e) > 0 {
		cloudInitUserData = e
	}
	if e := os.Getenv(metaDataEnv); len(e) > 0 {
		cloudInitMetaData = e
	}
	cloudInitOpt, err := generateCloudInitOpt(cloudinitConfig, cloudInitUserData, cloudInitMetaData)
	checkError(err)

	args := pflag.Args()
	log.SetLevel(log.DebugLevel)
	if len(args) == 0 {
		log.Fatalf("qemu launch command is required")
	}
	args = append(args, qemuNetworkOpt...)
	args = append(args, cloudInitOpt...)
	log.Infof("run qemu with command: %s", strings.Join(args, " "))
	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, syscall.SIGTERM, syscall.SIGINT)
	qemuCMD := exec.Command(args[0], args[1:]...)
	qemuCMD.Stdin = os.Stdin
	qemuCMD.Stdout = os.Stdout
	qemuCMD.Stderr = os.Stderr
	qemuCMD.ExtraFiles = []*os.File{tapFile}
	if err := qemuCMD.Start(); err != nil {
		log.Fatalf("failed to start qemu: %+v", err)
	}
	go func() {
		sig := <-exitSig
		log.Infof("recieve signal %+v", sig)
		_ = qemuCMD.Process.Kill()
	}()
	go func() {
		if len(consolePath) == 0 {
			return
		}
		var conn net.Conn
		for {
			stat, err := os.Stat(consolePath)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			if stat.Mode()&os.ModeSocket == 0 {
				return
			}
			conn, err = net.Dial("unix", consolePath)
			if err != nil {
				log.Errorf("failed to dial console socket: %+v", err)
				time.Sleep(time.Second)
				continue
			}
			break
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("failed to read from console socket: %+v", err)
				return
			}
			_, err = os.Stdout.Write(buf[:n])
			if err != nil {
				log.Errorf("failed to write to stdout: %+v", err)
			}
		}
	}()
	if err := qemuCMD.Wait(); err != nil {
		log.Errorf("failed to wait for qemu: %+v", err)
		return
	}
	log.Infof("qemu exited with code %d", qemuCMD.ProcessState.ExitCode())
}

func generateCloudinitConfig(nic *vmutil.NIC) (*cloudinit.NetworkConfig, error) {
	iPv4DefaultGateway, err := vmutil.GetIPv4DefaultGateway()
	if err != nil && !errors.Is(err, vmutil.NotFoundError) {
		return nil, errors.WithMessage(err, "failed to get default gateway")
	}
	iPv6DefaultGateway, err := vmutil.GetIPv6DefaultGateway()
	if err != nil && !errors.Is(err, vmutil.NotFoundError) {
		return nil, errors.WithMessage(err, "failed to get v6 default gateway")
	}
	addrs, err := nic.Addrs()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get addrs")
	}
	addresses := make([]*net.IPNet, 0)
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		addresses = append(addresses, ipNet)
	}
	return &cloudinit.NetworkConfig{
		Mac:       nic.HardwareAddr,
		Addresses: addresses,
		Gateway4:  iPv4DefaultGateway,
		Gateway6:  iPv6DefaultGateway,
	}, nil
}

func generateCloudInitOpt(n *cloudinit.NetworkConfig, userData string, metaData string) ([]string, error) {
	content, err := GenerateNetworkConfig(n)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to generate network config")
	}
	tempDir, err := os.MkdirTemp("", "cloud-init-*")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create temp dir")
	}
	err = os.WriteFile(filepath.Join(tempDir, "network-config"), content, os.ModePerm)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to write network-config")
	}
	err = os.WriteFile(filepath.Join(tempDir, "user-data"), []byte(userData), os.ModePerm)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to write user-data")
	}
	err = os.WriteFile(filepath.Join(tempDir, "meta-data"), []byte(metaData), os.ModePerm)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to write meta-data")
	}
	isoFile := "seed.iso"
	err = GenISO(tempDir, isoFile, []string{"network-config", "meta-data", "user-data"}, "cidata")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to generate cloud-init iso")
	}
	return []string{"-drive", fmt.Sprintf("driver=raw,file=%s,if=virtio", filepath.Join(tempDir, isoFile))}, nil
}

func generateQEMUNetworkOpt(fd int, macAddr net.HardwareAddr, mtu int) []string {
	return []string{
		"-netdev", fmt.Sprintf("tap,id=net0,vhost=on,fd=%d", fd),
		"-device", "virtio-net-pci,netdev=net0,mac=" + macAddr.String() + ",host_mtu=" + strconv.Itoa(mtu),
	}
}

func checkError(err error, s ...string) {
	if err == nil {
		return
	}
	log.Error(s)
	panic(err)
}

func GenISO(workDir string, output string, files []string, label string) error {
	args := []string{"-output", output, "-volid", label, "-joliet", "-rock"}
	args = append(args, files...)
	cmd := exec.Command("genisoimage", args...)
	cmd.Dir = workDir
	return cmd.Run()
}
