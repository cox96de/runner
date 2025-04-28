package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var runtimeImage string

func main() {
	root := &cobra.Command{}
	flagSet := root.Flags()
	flagSet.StringVarP(&runtimeImage, "runtime-image", "", "cox96de/runner-vm-runtime", "runtime image")
	root.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "run a image, use for install, configure image",
		Run: func(cmd *cobra.Command, args []string) {
			var cdroms []string
			flagSet := cmd.PersistentFlags()
			flagSet.StringArrayVarP(&cdroms, "cdrom", "", []string{}, "cdrom to use")
			err := flagSet.Parse(args)
			checkError(err)
			image := flagSet.Args()
			err = runImage(&runImageOption{
				CDRoms:   cdroms,
				Image:    image[0],
				Snapshot: "off",
			})
			checkError(err)
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "test a image with read only mode, bootstrap executor",
		Run: func(cmd *cobra.Command, args []string) {
			err := flagSet.Parse(args)
			checkError(err)
			image := flagSet.Args()
			err = runImage(&runImageOption{
				Image:    image[0],
				Snapshot: "on",
			})
			checkError(err)
		},
	})
	checkError(root.Execute())
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type runImageOption struct {
	CDRoms   []string
	Image    string
	Snapshot string
}

func runImage(option *runImageOption) error {
	qemuArgs := []string{
		"qemu-system-x86_64",
		"-nodefaults",
		"--nographic",
		"-display", "none",
		"-machine", "type=pc,usb=off",
		"-cpu", "host",
		"--enable-kvm",
		"-smp", "4,sockets=1,cores=4,threads=1",
		"-m", "4096M", "-device", "virtio-balloon-pci,id=balloon0", "-drive",
		fmt.Sprintf("file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=%s", option.Image, option.Snapshot),
		"-serial", "chardev:serial0", "-chardev", "socket,id=serial0,path=/work/console.sock,server=on,wait=off",
		"-vnc", "unix:/work/vnc.sock",
		"-device", "VGA",
	}
	for idx, s := range option.CDRoms {
		qemuArgs = append(qemuArgs, "-drive", fmt.Sprintf("file=%s,format=raw,if=ide,media=cdrom,index=%d", s, idx))
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dockerCMD := []string{"run", "--rm", "-v", wd + ":/work", "-w", "/work", "--privileged", runtimeImage, "--"}
	dockerCMD = append(dockerCMD, qemuArgs...)
	return run("docker", dockerCMD...)
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
