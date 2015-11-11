package main

import (
	"fmt"

	"github.com/megamsys/urknall"
)

type Hostname struct {
	Hostname string `urknall:"required=true"`
}

func (h *Hostname) Render(pkg urknall.Package) {
	pkg.AddCommands("base",
		Shell("hostname localhost"), // Set hostname to make sudo happy.
		&FileCommand{Path: "/etc/hostname", Content: h.Hostname},
		&FileCommand{Path: "/etc/hosts", Content: "127.0.0.1 {{ .Hostname }} localhost"},
		Shell("hostname -F /etc/hostname"),
	)
}

type System struct {
	Timezone string

	// limits
	LimitsDefaults bool

	// sysctl
	SysctlDefaults bool
	ShmMax         string
	ShmAll         string

	SwapInMB int
}

const TimezoneUTC = "Etc/UTC"

func (tpl *System) Render(pkg urknall.Package) {
	if tpl.Timezone != "" {
		pkg.AddCommands("timezone",
			WriteFile("/etc/timezone", tpl.Timezone, "root", 0644), // see TimezoneUTC
			Shell("dpkg-reconfigure --frontend noninteractive tzdata"),
		)
	}

	if tpl.SysctlDefaults {
		pkg.AddCommands("sysctl",
			WriteFile("/etc/sysctl.conf", sysctlTpl, "root", 0644),
			Shell("sysctl -p"),
		)
	}

	if tpl.LimitsDefaults {
		pkg.AddCommands("limits",
			WriteFile("/etc/security/limits.conf", limitsTpl, "root", 0644),
			Shell("ulimit -a"),
		)
	}

	if tpl.SwapInMB > 0 {
		pkg.AddCommands("swap",
			Shell(fmt.Sprintf("swapoff -a && rm -f /swapfile && fallocate -l %dM /swapfile", tpl.SwapInMB)),
			Shell("chmod 0600 /swapfile"),
			Shell("mkswap /swapfile"),
			Shell("grep '/swapfile' /etc/fstab > /dev/null || echo '/swapfile none swap defaults 0 0' >> /etc/fstab"),
			Shell("swapon -a"),
		)
	}
}

const limitsTpl = `* soft nofile 65535
* hard nofile 65535
root soft nofile 65535
root hard nofile 65535
`

const sysctlTpl = `net.core.rmem_max=16777216
net.core.wmem_max=16777216
net.core.wmem_default=262144
net.ipv4.tcp_rmem=4096 87380 16777216
net.ipv4.tcp_wmem=4096 65536 16777216
net.core.netdev_max_backlog=4000
net.ipv4.tcp_low_latency=1
net.ipv4.tcp_window_scaling=1
net.ipv4.tcp_timestamps=1
net.ipv4.tcp_sack=1
fs.file-max=65535
net.core.wmem_default=8388608
net.core.rmem_default=8388608
net.core.netdev_max_backlog=10000
net.core.somaxconn=4000
net.ipv4.tcp_max_syn_backlog=40000
net.ipv4.tcp_fin_timeout=15
net.ipv4.tcp_tw_reuse=1
vm.swappiness=0
{{ if .ShmMax }}kernel.shmmax={{ .ShmMax }}{{ end }}
{{ if .ShmAll }}kernel.shmmax={{ .ShmAll }}{{ end }}
`
