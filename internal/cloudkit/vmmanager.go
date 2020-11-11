package cloudkit

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/lithammer/shortuuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	libvirtxml "libvirt.org/libvirt-go-xml"
)

// VM represents a VM as understood by cloudkit
type VM struct {
	ID         int                         `json:"id,omitempty"`
	Name       string                      `json:"name,omitempty"`
	State      string                      `json:"state"`
	IP         string                      `json:"ip,omitempty"`
	MAC        string                      `json:"mac,omitempty"`
	Mem        int                         `json:"mem,omitempty"`
	CurrentMem int                         `json:"current_mem,omitempty"`
	VCPUs      int                         `json:"vcpus,omitempty"`
	Type       libvirtxml.DomainOSType     `json:"type,omitempty"`
	Devices    libvirtxml.DomainDeviceList `json:"devices,omitempty"`
}

// VMController describes all the actions you can take on a VM.
type VMController interface {
	CreateVM(machineType string, memoryInGB int, vCPUs int) error
	GetRunningVMs() ([]VM, error)
}

// VMManager imlements the VMController interface and handles
// everything to do with managing VMs in the hardware pool.
type VMManager struct {
	libvirt *libvirt.Libvirt
	logger  *logrus.Logger
}

// NewVMManager creates a tcp connection to libvirt on the host machines.
func NewVMManager(hostLibvirtConnStr string, log *logrus.Logger) (*VMManager, error) {
	protocol := "tcp"
	c, err := net.DialTimeout(protocol, hostLibvirtConnStr, 2*time.Second)
	if err != nil {
		return nil, err
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, err
	}

	v, err := l.Version()
	if err != nil {
		return nil, err
	}
	log.Infof("current libvirt version: %s\n\n", v)

	return &VMManager{libvirt: l, logger: log}, nil
}

// GetRunningVMs asks libvirt for current domains and returns them.
func (v *VMManager) GetRunningVMs() ([]VM, error) {
	net, err := v.libvirt.NetworkLookupByName("default")
	if err != nil {
		return nil, err
	}

	dms, err := v.libvirt.Domains()
	if err != nil {
		return nil, err
	}

	var runningVMs []VM
	for i := range dms {
		if dms[i].ID != -1 {
			rXML, err := v.libvirt.DomainGetXMLDesc(dms[i], 0)
			if err != nil {
				return nil, err
			}

			domcfg := &libvirtxml.Domain{}
			if err := domcfg.Unmarshal(rXML); err != nil {
				return nil, err
			}

			var macAddr string
			for _, val := range domcfg.Devices.Interfaces {
				if val.Source.Network.Network == "default" {
					macAddr = val.MAC.Address
				}
			}
			if macAddr == "" {
				macAddr = "pending"
			}

			ip := "pending"
			if macAddr != "pending" {
				m := libvirt.OptString{macAddr}
				leases, _, err := v.libvirt.NetworkGetDhcpLeases(net, m, 1, 0)
				if err != nil {
					return nil, err
				}
				for _, val := range leases {
					if len(val.Mac) == 0 {
						continue
					} else if val.Mac[0] == macAddr {
						ip = val.Ipaddr
					}
				}
			}

			st, maxMem, mem, rNrVirtCPU, cpuTime, err := v.libvirt.DomainGetInfo(dms[i])
			fmt.Println(st, maxMem, mem, rNrVirtCPU, cpuTime, err)

			state, _, err := v.libvirt.DomainGetState(dms[i], 0)
			if err != nil {
				return nil, err
			}

			vm := VM{
				ID:    int(dms[i].ID),
				Name:  dms[i].Name,
				State: domainState(state),
				// TODO once I figure out networking they will get their own IP...
				// For now we have to search for the dhcp lease above and strip it.
				IP:         ip,
				MAC:        macAddr,
				Mem:        int(domcfg.Memory.Value),
				CurrentMem: int(domcfg.CurrentMemory.Value),
				VCPUs:      int(domcfg.VCPU.Value),
				Type:       *domcfg.OS.Type,
				Devices:    *domcfg.Devices,
			}
			runningVMs = append(runningVMs, vm)
		}
	}

	return runningVMs, nil
}

// CreateVM currently handles spinning up the default ubuntu bionic VM
func (v *VMManager) CreateVM(machineType string, memoryInGB int, vCPUs int) error {
	id := shortuuid.New()

	pk, err := pubKey(os.Getenv("PUBKEY_LOCATION"))
	if err != nil {
		return err
	}

	if err := prepareHostWithUbuntuDisks(pk, id); err != nil {
		return err
	}

	b, err := xml.Marshal(buildDomainXML(id, machineType, memoryInGB, vCPUs))
	if err != nil {
		return err
	}

	domain, err := v.libvirt.DomainCreateXML(string(b), 0)
	if err != nil {
		return err
	}
	fmt.Printf("domain: %+v", domain)

	return err
}

func domainState(state int32) string {
	switch state {
	case int32(libvirt.DomainNostate):
		return "unknown"
	case int32(libvirt.DomainRunning):
		return "running"
	case int32(libvirt.DomainBlocked):
		return "blocked"
	case int32(libvirt.DomainPaused):
		return "paused"
	case int32(libvirt.DomainShutdown):
		return "shutting down"
	case int32(libvirt.DomainShutoff):
		return "off"
	case int32(libvirt.DomainCrashed):
		return "crashed"
	case int32(libvirt.DomainPmsuspended):
		return "pm suspended"
	default:
		return "unknown"
	}
}

func pubKey(path string) (ssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func prepareHostWithUbuntuDisks(pk ssh.AuthMethod, id string) error {
	user := "root"
	remote := "157.245.225.232"
	port := ":22"
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{pk},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", remote+port, config)
	if err != nil {
		return err
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// Create a single command that is semicolon seperated
	commands := []string{
		"cp /var/lib/libvirt/images/ubuntu-bionic.img /var/lib/libvirt/images/ubuntu-bionic-" + id + ".img",
		"cp /var/lib/libvirt/images/ubuntu-bionic.iso /var/lib/libvirt/images/ubuntu-bionic-" + id + ".iso",
		"cloud-localds /var/lib/libvirt/images/ubuntu-bionic-" + id + ".iso cloud.txt",
	}
	combined := strings.Join(commands, "; ")

	if err := sess.Run(combined); err != nil {
		return err
	}

	return nil
}

// Currently builds an ubuntu 18.04 bionic beaver image with user defined memroy and cpus.
func buildDomainXML(id string, machineType string, memoryInGB int, numVCPUs int) libvirtxml.Domain {
	return libvirtxml.Domain{
		Type: "kvm",
		Name: "ubuntu-bionic-" + id,
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{Type: "hvm"},
		},
		Memory: &libvirtxml.DomainMemory{
			Unit:  "MiB",
			Value: gbToMiB(memoryInGB),
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     vCPUCount(numVCPUs),
		},
		Devices: &libvirtxml.DomainDeviceList{
			Interfaces: []libvirtxml.DomainInterface{{
				Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
				Source: &libvirtxml.DomainInterfaceSource{
					Network: &libvirtxml.DomainInterfaceSourceNetwork{Network: "default"},
				},
			}},
			Disks: []libvirtxml.DomainDisk{{
				Driver: &libvirtxml.DomainDiskDriver{Type: "raw"},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: "/var/lib/libvirt/images/ubuntu-bionic-" + id + ".img",
					},
				},
				Target: &libvirtxml.DomainDiskTarget{Dev: "vda", Bus: "virtio"},
			}, {
				Driver: &libvirtxml.DomainDiskDriver{Type: "raw"},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: "/var/lib/libvirt/images/ubuntu-bionic-" + id + ".iso",
					},
				},
				Target: &libvirtxml.DomainDiskTarget{Dev: "hdc", Bus: "ide"},
			}},
		},
	}
}

func gbToMiB(gb int) uint {
	switch gb {
	case 1:
		return 1024
	case 2:
		return 2048
	case 4:
		return 4096
	case 8:
		return 8192
	default:
		return 2048
	}
}

func vCPUCount(numVCPUs int) uint {
	switch numVCPUs {
	case 1:
		return 1
	case 2:
		return 2
	case 4:
		return 4
	default:
		return 1
	}
}
