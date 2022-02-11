package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()

	default:
		panic("bad command")
	}
}

func run() {
	fmt.Printf("Running %v as user %d in process %d\n", os.Args[2:], os.Geteuid(), os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// CLONE_NEWUTS: unix timesharing system -> decouple hostname
		// CLONE_NEWPID: process ID
		// CLONE_NEWNS: mount
		// CLONE_NEWUSER: create new user ns with full capabilities in the new ns
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		// Unshare mount with the host
		Unshareflags: syscall.CLONE_NEWNS,
		// Map process ID 1000 to container ID 0 (?)
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      1000,
			Size:        1,
		}},
	}
	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as user %d in process %d\n", os.Args[2:], os.Geteuid(), os.Getpid())

	createCGroup()

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("./ubuntufs"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
}

func createCGroup() {
	cgroups := "/sys/fs/cgroup"
	pids := filepath.Join(cgroups, "pids")
	err := os.Mkdir(filepath.Join(pids, "liz"), 0o755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	// Set max number of processes allowed in container
	must(os.WriteFile(filepath.Join(pids, "liz/pids.max"), []byte("20"), 0o700))
	// Remove the new cgroup after the container exits
	must(os.WriteFile(filepath.Join(pids, "liz/notify_on_release"), []byte("1"), 0o700))
	// Add current process id to cgroup
	must(os.WriteFile(filepath.Join(pids, "liz/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0o700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
