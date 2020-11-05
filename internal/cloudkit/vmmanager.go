package cloudkit

import (
	"encoding/xml"
	"fmt"
	"net"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/sirupsen/logrus"
	libvirtxml "libvirt.org/libvirt-go-xml"
)

// VM ...
type VM struct{}

// VMController describes all the actions you can take on a VM.
type VMController interface {
	CreateVM() error
	GetVMs() ([]libvirt.Domain, error)
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

// CreateVM ...
func (v *VMManager) CreateVM() error {
	b, err := xml.Marshal(defaultDomain())
	if err != nil {
		fmt.Println("err1", err)
		return err
	}

	domain, err := v.libvirt.DomainCreateXML(string(b), 0)
	if err != nil {
		fmt.Println("err2", err)
		return err
	}
	fmt.Printf("domain: %+v", domain)

	return err
}

// GetVMs ...
func (v *VMManager) GetVMs() ([]libvirt.Domain, error) {
	dms, err := v.libvirt.Domains()
	if err != nil {
		return nil, err
	}
	return dms, nil
}

// Creates a domain definition with some defaults
func defaultDomain() libvirtxml.Domain {
	return libvirtxml.Domain{
		Type: "kvm",
		Name: "ubuntu-bionic",
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{Type: "hvm"},
		},
		Memory: &libvirtxml.DomainMemory{
			Unit:  "MiB",
			Value: 1024,
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     1,
		},
		CPU: &libvirtxml.DomainCPU{},
		Devices: &libvirtxml.DomainDeviceList{
			// Emulator: "/usr/bin/qemu-system-x86_64",
			// Graphics: []libvirtxml.DomainGraphic{{
			// 	VNC: &libvirtxml.DomainGraphicVNC{},
			// }},
			Graphics: []libvirtxml.DomainGraphic{},
			Interfaces: []libvirtxml.DomainInterface{{
				Model: &libvirtxml.DomainInterfaceModel{
					Type: "virtio",
				},
				Source: &libvirtxml.DomainInterfaceSource{
					Network: &libvirtxml.DomainInterfaceSourceNetwork{
						Network: "default",
					},
				},
				// MAC: &libvirtxml.DomainInterfaceMAC{
				// 	Address: "4a:ee:c7:15:81:d3",
				// },
				// IP: []libvirtxml.DomainInterfaceIP{{
				// 	Address: "192.168.122.10",
				// 	// Address: "192.168.122.2",
				// }},
			}},
			Disks: []libvirtxml.DomainDisk{{
				Driver: &libvirtxml.DomainDiskDriver{
					Type: "raw",
				},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: "/var/lib/libvirt/images/ubuntu-bionic.img",
					},
				},
				Target: &libvirtxml.DomainDiskTarget{
					Dev: "vda",
					Bus: "virtio",
				},
			}, {
				Driver: &libvirtxml.DomainDiskDriver{
					Type: "raw",
				},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: "/var/lib/libvirt/images/ubuntu-bionic.iso",
					},
				},
				Target: &libvirtxml.DomainDiskTarget{
					Dev: "hdc",
					Bus: "ide",
				},
			}},
		},
	}
}
