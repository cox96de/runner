//go:build vm_runtime_integration

package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cox96de/containervm/util"

	"github.com/cox96de/runner/app/executor/executorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gotest.tools/v3/fs"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/testtool"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
)

func TestRunCMD(t *testing.T) {
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	env.ChangeWorkingDir(t, gitRoot)
	containerName := "vm-runtime-test-run-cmd"
	_ = run("docker", "rm", containerName)
	err = runtimeBuild()
	assert.NilError(t, err)
	imagePath := "engine/vm/runtime/debian-11.qcow2"
	_ = imagePath
	metaData := `instance-id: vm-runner
local-hostname: vm-runner
`
	userData := `#cloud-config

runcmd:
  - [sh, -c, "nohup python3 -m http.server > /var/server.log 2>&1 &"]`
	qemuCMD := fmt.Sprintf(
		"--console /tmp/console.sock " +
			"-- " +
			"qemu-system-x86_64 " +
			"-nodefaults " +
			"--nographic " +
			"-display none " +
			"-machine type=pc,usb=off " +
			"-cpu host " +
			"--enable-kvm " +
			"-smp 4,sockets=1,cores=4,threads=1 " +
			"-m 4096M -device virtio-balloon-pci,id=balloon0 " +
			fmt.Sprintf("-drive file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=on ", imagePath) +
			"-serial chardev:serial0 -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off " +
			"-vnc unix:/tmp/vnc.sock -device VGA ",
	)
	dockerRunCMD := "docker run " +
		"--privileged " +
		fmt.Sprintf("-e CLOUD_INIT_USER_DATA='%s' ", userData) +
		fmt.Sprintf("-e CLOUD_INIT_META_DATA='%s' ", metaData) +
		"-v /tmp/containervm:/tmp " +
		"-v $PWD:/root " +
		"--entrypoint='' " +
		"--name " + containerName + " " +
		"-w /root " +
		runtimeImage + " " +
		runtimeBinary + " " +
		qemuCMD
	dockerProcessChan := make(chan error)
	go func() {
		err := run("bash", "-c", dockerRunCMD)
		if err != nil {
			t.Logf("qemu image exit with: %+v", err)
			dockerProcessChan <- err
		}
	}()
	defer func() {
		_ = run("docker", "stop", containerName)
		_ = run("docker", "rm", containerName)
	}()

	var ip string
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		time.Sleep(time.Second * 3)
		ip, err = getContainerIP(containerName)
		assert.NilError(t, err)
		if ip != "" {
			break
		}
	}
	ip = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(ip), "'"), "'")
	t.Logf("ip: %s", ip)
	testVM := func() (string, error) {
		request, err := http.NewRequest(http.MethodGet, "http://"+ip+":8000", nil)
		if err != nil {
			return "", err
		}
		// Use custom client to disable proxy setting.
		client := &http.Client{
			Transport:     &http.Transport{},
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}
		resp, err := client.Do(request)
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != 200 {
			return string(body), errors.Errorf("status code: %d", resp.StatusCode)
		}
		t.Logf("output: %s", string(body))
		return string(body), nil
	}
	pass := false
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		output, err := testVM()
		if err == nil {
			pass = true
			break
		}
		t.Logf("error: %+v, output: %s", err, output)
		time.Sleep(time.Second * 5)
	}
	assert.Assert(t, pass)
}

func TestRunWindows(t *testing.T) {
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	env.ChangeWorkingDir(t, gitRoot)
	containerName := "vm-runtime-test-run-windows"
	_ = run("docker", "rm", containerName)
	err = runtimeBuild()
	assert.NilError(t, err)
	err = run("bash", "-c", "GOOS=windows go build -o output/executor.exe ./cmd/executor")
	assert.NilError(t, err)
	err = run("bash", "-c", "GOOS=windows go build -o output/executor.exe ./cmd/executor")
	assert.NilError(t, err)
	err = run("bash", "-c", "genisoimage -output output/executor.iso -joliet -rock output/executor.exe engine/vm/runtime/windows_boot.ps1")
	assert.NilError(t, err)
	imagePath := "engine/vm/runtime/windows.qcow2"
	_ = imagePath
	metaData := `instance-id: vm-runner
local-hostname: vm-runner
`
	userData := `#cloud-config

runcmd:
  - [powershell,"D:\\windows_boot.ps1"]
  - [powershell,"E:\\windows_boot.ps1"]
  - [powershell,"F:\\windows_boot.ps1"]
  - [powershell,"G:\\windows_boot.ps1"]
  - [powershell,"Z:\\windows_boot.ps1"]
`
	qemuCMD := fmt.Sprintf(
		"-- " +
			"qemu-system-x86_64 " +
			"-nodefaults " +
			"--nographic " +
			"-display none " +
			"-machine type=pc,usb=off " +
			"-cpu host " +
			"--enable-kvm " +
			"-smp 4,sockets=1,cores=4,threads=1 " +
			"-m 4096M -device virtio-balloon-pci,id=balloon0 " +
			fmt.Sprintf("-drive file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=on ", imagePath) +
			"-cdrom output/executor.iso " +
			"-serial chardev:serial0 -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off " +
			"-vnc unix:/tmp/vnc.sock -device VGA ",
	)
	dockerRunCMD := "docker run " +
		"--privileged " +
		fmt.Sprintf("-e CLOUD_INIT_USER_DATA='%s' ", userData) +
		fmt.Sprintf("-e CLOUD_INIT_META_DATA='%s' ", metaData) +
		"-v /tmp/containervm:/tmp " +
		"-v $PWD:/root " +
		"--entrypoint='' " +
		"--name " + containerName + " " +
		"-w /root " +
		runtimeImage + " " +
		runtimeBinary + " " +
		qemuCMD
	dockerProcessChan := make(chan error)
	go func() {
		err := run("bash", "-c", dockerRunCMD)
		if err != nil {
			t.Logf("qemu image exit with: %+v", err)
			dockerProcessChan <- err
		}
	}()
	defer func() {
		_ = run("docker", "stop", containerName)
		_ = run("docker", "rm", containerName)
	}()

	var ip string
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		time.Sleep(time.Second * 3)
		ip, err = getContainerIP(containerName)
		assert.NilError(t, err)
		if ip != "" {
			break
		}
	}
	ip = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(ip), "'"), "'")
	t.Logf("ip: %s", ip)
	testVM := func() (string, error) {
		conn, err := grpc.NewClient(ip+":8080", grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithNoProxy())
		if err != nil {
			return "", errors.WithMessage(err, "failed to connect to executor")
		}
		client := executorpb.NewExecutorClient(conn)
		_, err = client.Ping(context.Background(), &executorpb.PingRequest{})
		if err != nil {
			return "", errors.WithMessage(err, "failed to ping executor")
		}
		environment, err := client.Environment(context.Background(), &executorpb.EnvironmentRequest{})
		if err != nil {
			return "", errors.WithMessage(err, "failed to ping executor")
		}
		t.Logf("%+v", environment.Environment)
		return "", nil
	}
	pass := false
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		output, err := testVM()
		if err == nil {
			pass = true
			break
		}
		t.Logf("error: %+v, output: %s", err, output)
		time.Sleep(time.Second * 5)
	}
	assert.Assert(t, pass)
}

func TestMount9P(t *testing.T) {
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	env.ChangeWorkingDir(t, gitRoot)
	containerName := "vm-runtime-test-mount-9p"
	_ = run("docker", "rm", containerName)
	err = runtimeBuild()
	assert.NilError(t, err)
	content := "hello world 9p"
	testDir := fs.NewDir(t, "mount9p", fs.WithFile("index.html", content))
	imagePath := "engine/vm/runtime/debian-11.qcow2"
	metaData := `instance-id: vm-runner
local-hostname: vm-runner
`
	userData := `#cloud-config

runcmd:
  - [sh, -c, "nohup python3 -m http.server > /var/server.log 2>&1 &"]
mounts:
  - [9ptest, /mnt/9p, "9p", "defaults", "0", "0"]
`
	qemuCMD := fmt.Sprintf(
		"-- " +
			"qemu-system-x86_64 " +
			"-nodefaults " +
			"--nographic " +
			"-display none " +
			"-machine type=pc,usb=off " +
			"-cpu host " +
			"--enable-kvm " +
			"-smp 4,sockets=1,cores=4,threads=1 " +
			"-m 4096M -device virtio-balloon-pci,id=balloon0 " +
			"-fsdev local,security_model=passthrough,id=fsdev0,path=/mnt/9p " +
			"-device virtio-9p-pci,fsdev=fsdev0,mount_tag=9ptest " +
			fmt.Sprintf("-drive file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=on ", imagePath) +
			"-serial chardev:serial0 -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off " +
			"-vnc unix:/tmp/vnc.sock -device VGA ",
	)
	dockerRunCMD := "docker run " +
		"--privileged " +
		fmt.Sprintf("-e CLOUD_INIT_USER_DATA='%s' ", userData) +
		fmt.Sprintf("-e CLOUD_INIT_META_DATA='%s' ", metaData) +
		"-v /tmp/containervm:/tmp " +
		"-v $PWD:/root " +
		"--entrypoint='' " +
		"-v " + testDir.Path() + ":/mnt/9p " +
		"--name " + containerName + " " +
		"-w /root " +
		runtimeImage + " " +
		runtimeBinary + " " +
		qemuCMD
	dockerProcessChan := make(chan error)
	go func() {
		err := run("bash", "-c", dockerRunCMD)
		if err != nil {
			t.Logf("qemu image exit with: %+v", err)
			dockerProcessChan <- err
		}
	}()
	defer func() {
		_ = run("docker", "stop", containerName)
		_ = run("docker", "rm", containerName)
	}()

	var ip string
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		time.Sleep(time.Second * 3)
		ip, err = getContainerIP(containerName)
		assert.NilError(t, err)
		if ip != "" {
			break
		}
	}
	ip = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(ip), "'"), "'")
	t.Logf("ip: %s", ip)
	testVM := func() (string, error) {
		request, err := http.NewRequest(http.MethodGet, "http://"+ip+":8000/mnt/9p/index.html", nil)
		if err != nil {
			return "", err
		}
		// Use custom client to disable proxy setting.
		client := &http.Client{
			Transport:     &http.Transport{},
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}
		resp, err := client.Do(request)
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != 200 {
			return string(body), errors.Errorf("status code: %d", resp.StatusCode)
		}
		t.Logf("output: %s", string(body))
		return string(body), nil
	}
	pass := false
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		output, err := testVM()
		if err == nil {
			pass = true
			break
		}
		t.Logf("error: %+v, output: %s", err, output)
		time.Sleep(time.Second * 5)
	}
	assert.Assert(t, pass)
}

func TestRunExecutor(t *testing.T) {
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	env.ChangeWorkingDir(t, gitRoot)
	containerName := "vm-runtime-test-run-executor"
	_ = run("docker", "rm", containerName)
	err = runtimeBuild()
	assert.NilError(t, err)
	err = run("bash", "-c", "CGO_ENABLED=0 go build -o output/executor ./cmd/executor")
	assert.NilError(t, err)
	imagePath := "engine/vm/runtime/debian-11.qcow2"
	metaData := `instance-id: vm-runner
local-hostname: vm-runner
`
	userData := `#cloud-config

runcmd:
  - [sh, -c, "while true; do if [ -f /mnt/9p/executor ]; then nohup /mnt/9p/executor > /var/server.log 2>&1 & break; else echo \"File not found. Retrying in 1 second...\"; sleep 1; fi; done"]
mounts:
  - [9ptest, /mnt/9p, "9p", "trans=virtio,version=9p2000.L,msize=104857600", "0", "0"]
`
	qemuCMD := fmt.Sprintf(
		"-- " +
			"qemu-system-x86_64 " +
			"-nodefaults " +
			"--nographic " +
			"-display none " +
			"-machine type=pc,usb=off " +
			"-cpu host " +
			"--enable-kvm " +
			"-smp 4,sockets=1,cores=4,threads=1 " +
			"-m 4096M -device virtio-balloon-pci,id=balloon0 " +
			"-fsdev local,security_model=passthrough,id=fsdev0,path=/mnt/9p " +
			"-device virtio-9p-pci,fsdev=fsdev0,mount_tag=9ptest " +
			fmt.Sprintf("-drive file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=on ", imagePath) +
			"-serial chardev:serial0 -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off " +
			"-vnc unix:/tmp/vnc.sock -device VGA ",
	)
	dockerRunCMD := "docker run " +
		"--privileged " +
		fmt.Sprintf("-e CLOUD_INIT_USER_DATA='%s' ", userData) +
		fmt.Sprintf("-e CLOUD_INIT_META_DATA='%s' ", metaData) +
		"-v /tmp/containervm:/tmp " +
		"-v $PWD:/root " +
		"--entrypoint='' " + "-v " + "$PWD/output:/mnt/9p " +
		"--name " + containerName + " " +
		"-w /root " +
		runtimeImage + " " +
		runtimeBinary + " " +
		qemuCMD
	dockerProcessChan := make(chan error)
	go func() {
		err := run("bash", "-c", dockerRunCMD)
		if err != nil {
			t.Logf("qemu image exit with: %+v", err)
			dockerProcessChan <- err
		}
	}()
	defer func() {
		_ = run("docker", "stop", containerName)
		_ = run("docker", "rm", containerName)
	}()

	var ip string
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		time.Sleep(time.Second * 3)
		ip, err = getContainerIP(containerName)
		assert.NilError(t, err)
		if ip != "" {
			break
		}
	}
	ip = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(ip), "'"), "'")
	t.Logf("ip: %s", ip)
	testVM := func() (string, error) {
		conn, err := grpc.NewClient(ip+":8080", grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithNoProxy())
		if err != nil {
			return "", errors.WithMessage(err, "failed to connect to executor")
		}
		client := executorpb.NewExecutorClient(conn)
		_, err = client.Ping(context.Background(), &executorpb.PingRequest{})
		if err != nil {
			return "", errors.WithMessage(err, "failed to ping executor")
		}
		return "", nil
	}
	pass := false
	for i := 0; i < 100; i++ {
		select {
		case err := <-dockerProcessChan:
			assert.NilError(t, err)
			t.FailNow()
		default:

		}
		output, err := testVM()
		if err == nil {
			pass = true
			break
		}
		t.Logf("error: %+v, output: %s", err, output)
		time.Sleep(time.Second * 5)
	}
	assert.Assert(t, pass)
}

const (
	runtimeImage  = "cox96de/runner-vm-runtime:latest"
	runtimeBinary = "engine/vm/runtime/runtime"
)

var runtimeBuildOnce = sync.Once{}

func runtimeBuild() error {
	var err error
	runtimeBuildOnce.Do(func() {
		err = run("bash", "-c", fmt.Sprintf("CGO_ENABLED=0 go build -o %s ./engine/vm/runtime/", runtimeBinary))
	})
	return err
}

func run(command string, args ...string) (err error) {
	fmt.Printf("run command: %s %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.WithMessage(err, "failed to run command")
	}
	if exitCode := cmd.ProcessState.ExitCode(); exitCode != 0 {
		return errors.Errorf("command exited with code %d", exitCode)
	}
	return nil
}

func getContainerIP(containerName string) (string, error) {
	return util.Run("docker", "inspect", "-f",
		"'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'", containerName)
}
