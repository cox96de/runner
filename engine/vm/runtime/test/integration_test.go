//go:build vm_runtime_integration

package test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/cox96de/containervm/util"
	"github.com/cox96de/runner/testtool"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

const imageName = "vm-runtime"

func dockerBuild() error {
	return run("docker", "build", "-t", imageName, "-f", "engine/vm/runtime/Dockerfile", ".")
}

func TestRuntime(t *testing.T) {
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	env.ChangeWorkingDir(t, gitRoot)
	containerName := "vm-runtime-test"
	_ = run("docker", "rm", containerName)
	err = dockerBuild()
	assert.NilError(t, err)
	imagePath := "engine/vm/runtime/debian-11-amd64.qcow2"
	_ = imagePath
	cloudInitUserDataEnv := "CLOUD_INIT_USER_DATA"
	cloudInitMetaDataEnv := "CLOUD_INIT_META_DATA"
	metaData := `instance-id: vm-runner
local-hostname: vm-runner
`
	userData := `#cloud-config

users:
- name: newsuper
  gecos: Big Stuff
  groups: users, admin
  sudo: ALL=(ALL) NOPASSWD:ALL
  shell: /bin/bash
  lock_passwd: false
  plain_text_passwd: password
  ssh_authorized_keys: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDH/DwEnbEOapaUjzTXIfVX0W+zn5KZBAg7nTIRyHkpkC8WJtyn7AkbtdmdFBNUhRLLHvy33S7WkSSYG2Ch1wWQVGAD8q73F4U2tTErHcRyN6CzIrpY9plX7QowjRgyQK5uODvGZ5muImy3VbBD+PyPhn5g78gg2TXdL/8Zzd4C/qrPqMdMwyNQBYXF9ZI1O5EgkyfmKd0irigwkItEXRoJ0BIN+tO3Ag+gpHJLEkE2V+lwBDT7o8v+063XOJIzKcoKw3VOGO3adRZxP9ov0UQG69uklav6p43wx8b6wOwr0AvEnLLaoZTK5vRJhdWK9HRADsNYxCaKXb243a9Rz4oT k8sindocker`
	qemuCMD := fmt.Sprintf(
		fmt.Sprintf("--cloud-init-user-data-env=%s ", cloudInitUserDataEnv) +
			fmt.Sprintf("--cloud-init-meta-data-env=%s ", cloudInitMetaDataEnv) +
			"-- " +
			"qemu-system-x86_64 " +
			"-nodefaults " +
			"--nographic " +
			"-display none " +
			"-machine type=pc,usb=off " +
			"-smp 4,sockets=1,cores=4,threads=1 " +
			"-m 4096M -device virtio-balloon-pci,id=balloon0 " +
			fmt.Sprintf("-drive file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=on ", imagePath) +
			"-serial chardev:serial0 -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off " +
			"-vnc unix:/tmp/vnc.sock -device VGA ",
	)
	// TODO: add -enable-kvm if kvm is enabled.
	dockerRunCMD := "docker run " +
		"--privileged " +
		fmt.Sprintf("-e %s='%s' ", cloudInitUserDataEnv, userData) +
		fmt.Sprintf("-e %s='%s' ", cloudInitMetaDataEnv, metaData) +
		"-v /tmp/containervm:/tmp " +
		"-v $PWD:/root " +
		"--name " + containerName + " " +
		"-w /root " +
		imageName + " " +
		//"-- " +
		qemuCMD

	go func() {
		err := run("bash", "-c", dockerRunCMD)
		assert.Check(t, err != nil)
	}()
	defer func() {
		_ = run("docker", "stop", containerName)
		_ = run("docker", "rm", containerName)
	}()

	var ip string
	for i := 0; i < 100; i++ {
		time.Sleep(time.Second * 3)
		ip, err = getContainerIP(containerName)
		assert.NilError(t, err)
		if ip != "" {
			break
		}
	}
	ip = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(ip), "'"), "'")
	t.Logf("ip: %s", ip)
	privateKey := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAx/w8BJ2xDmqWlI801yH1V9Fvs5+SmQQIO50yEch5KZAvFibcp+wJ
G7XZnRQTVIUSyx78t90u1pEkmBtgodcFkFRgA/Ku9xeFNrUxKx3EcjegsyK6WPaZV+0KMI
0YMkCubjg7xmeZriJst1WwQ/j8j4Z+YO/IINk13S//Gc3eAv6qz6jHTMMjUAWFxfWSNTuR
IJMn5indIq4oMJCLRF0aCdASDfrTtwIPoKRySxJBNlfpcAQ0+6PL/tOt1ziSMynKCsN1Th
jt2nUWcT/aL9FEBuvbpJWr+qeN8MfG+sDsK9ALxJyy2qGUyub0SYXVivR0QA7DWMQmil29
uN2vUc+KEwAAA9CpZusiqWbrIgAAAAdzc2gtcnNhAAABAQDH/DwEnbEOapaUjzTXIfVX0W
+zn5KZBAg7nTIRyHkpkC8WJtyn7AkbtdmdFBNUhRLLHvy33S7WkSSYG2Ch1wWQVGAD8q73
F4U2tTErHcRyN6CzIrpY9plX7QowjRgyQK5uODvGZ5muImy3VbBD+PyPhn5g78gg2TXdL/
8Zzd4C/qrPqMdMwyNQBYXF9ZI1O5EgkyfmKd0irigwkItEXRoJ0BIN+tO3Ag+gpHJLEkE2
V+lwBDT7o8v+063XOJIzKcoKw3VOGO3adRZxP9ov0UQG69uklav6p43wx8b6wOwr0AvEnL
LaoZTK5vRJhdWK9HRADsNYxCaKXb243a9Rz4oTAAAAAwEAAQAAAQA33qLR01A0q9h3lm53
r7gAGbWwI+NrtjGqnebwCua2kt5kvOSmUQ3WXP53oLUpxqeScYy+vR8puJDVochkTlLymG
/ein0Q8NQ5jXM4DW/lTN8rTIds9S+v3bwcBj79Qw64IiOo8SaA/IMM0PaWdsfwPO2vnS12
59fhfFgzWE0u3oJBcjATHQhkpfyVXLI5QExQL7/1mz75OhxqZHqlNS8vtvNzL+lq5iPzVg
3e1M9w+s0KwQGeEQu71FUEUTeG+o0yvtTf0pwpolNNEv+c/Vg0AQ1IxMnYW74WtjZUwXgC
IqqUkNPhzu3+88yWog50pVZrrWmqE45FYqokk0+xM47pAAAAgQC1nQxDpfSVIbvF7HsDUE
6aPABF/jDzR69qs5HxA2zuKt3qK8FRFooY97gERzgNchTx4Fc7LvwjNgPsJ/16n9QK9Z6D
5Y8u7cvSB73bJxQbuBlBeXNSmsoDSkoBcH7AzGHLptlneJiIImOV8omHR+ZakcJpx8wJZ2
G+dE5C2eERNAAAAIEA7M0ze83HFl+dMllscJjqzW9go7sUAqoXuP9TYnU+D0cohMWcFg4y
grmkR8XK2vIelIIU8ij+Dgx3mHTaQjr5lcdoLT5dfJ7mhvsOezCnGynNt3ui9lcTvvMTFw
nAMqYEIz9aWCzN4IK0ytOV24oz5BCpOFJxLZ5mm+hlu4qLGn0AAACBANgy6g2M+q/rhb29
XOIo/z2Gt/JCVbrdWys3xSK+w7RvitutzkvcaUXS/ffPDa1a25SxyFRT2P1Q1iLhyO3FEX
uYmwgZBgmbcKAb4K8oJaeX7iZ7xxAA+/SnGIwpU27D8By2fBvAD/hgf/ceZlzRu6P71Sh8
PzJY8Yglizre5MvPAAAAFnhpYXppaGFvQE1hY0Jvb2subG9jYWwBAgME
-----END OPENSSH PRIVATE KEY-----
`
	sshPrivateDir := fs.NewDir(t, "gotest", fs.WithFile("ssh.private", privateKey, fs.WithMode(0o600)))
	testVM := func(commands ...string) (string, error) {
		output, err := util.Run("ssh", append([]string{
			"-i", sshPrivateDir.Join("ssh.private"), "-o",
			"StrictHostKeyChecking=no", "newsuper@" + ip,
		}, commands...)...)
		if err != nil {
			return "", err
		}
		t.Logf("output: %s", output)
		return output, nil
	}
	pass := false
	for i := 0; i < 100; i++ {
		output, err := testVM("date")
		if err == nil {
			pass = true
			break
		}
		t.Logf("error: %+v, output: %s", err, output)
		time.Sleep(time.Second * 5)
	}
	assert.Assert(t, pass)
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
