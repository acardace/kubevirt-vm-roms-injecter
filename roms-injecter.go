package main

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"

	vmSchema "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

//go:embed OVMF_VARS.fd
var ovmfVars []byte

//go:embed OVMF_VARS.secboot.fd
var ovmfVarsSecBoot []byte

//go:embed OVMF_CODE.fd
var ovmfCode []byte

//go:embed OVMF_CODE.secboot.fd
var ovmfCodeSecBoot []byte

const (
	ovmfVarsPath        = "/var/run/kubevirt-hooks/OVMF_VARS.fd"
	ovmfCodePath        = "/var/run/kubevirt-hooks/OVMF_CODE.fd"
	ovmfVarsSecBootPath = "/var/run/kubevirt-hooks/OVMF_VARS.secboot.fd"
	ovmfCodeSecBootPath = "/var/run/kubevirt-hooks/OVMF_CODE.secboot.fd"
)

func copyFiles() error {
	if err := os.WriteFile(ovmfVarsPath, ovmfVars, 0666); err != nil {
		return err
	}
	if err := os.WriteFile(ovmfCodePath, ovmfCode, 0666); err != nil {
		return err
	}
	if err := os.WriteFile(ovmfVarsSecBootPath, ovmfVarsSecBoot, 0666); err != nil {
		return err
	}
	if err := os.WriteFile(ovmfCodeSecBootPath, ovmfCodeSecBoot, 0666); err != nil {
		return err
	}
	return nil
}

func onDefineDomain(vmiJSON, domainXML []byte) (string, error) {
	vmiSpec := vmSchema.VirtualMachineInstance{}
	if err := json.Unmarshal(vmiJSON, &vmiSpec); err != nil {
		return "", err
	}

	domainSpec := api.DomainSpec{}
	if err := xml.Unmarshal(domainXML, &domainSpec); err != nil {
		return "", err
	}

	fw := vmiSpec.Spec.Domain.Firmware
	if fw != nil &&
		fw.Bootloader != nil &&
		fw.Bootloader.EFI != nil &&
		fw.Bootloader.EFI.SecureBoot != nil &&
		*fw.Bootloader.EFI.SecureBoot {
		domainSpec.OS.NVRam.Template = ovmfVarsSecBootPath
		domainSpec.OS.BootLoader.Path = ovmfCodeSecBootPath
	} else if fw != nil &&
		fw.Bootloader != nil &&
		fw.Bootloader.EFI != nil {
		domainSpec.OS.NVRam.Template = ovmfVarsPath
		domainSpec.OS.BootLoader.Path = ovmfCodePath
	}

	newDomainXML, err := xml.Marshal(domainSpec)
	if err != nil {
		return "", err
	}

	return string(newDomainXML), nil
}

func main() {
	var vmiJSON, domainXML string
	pflag.StringVar(&vmiJSON, "vmi", "", "VMI to change in JSON format")
	pflag.StringVar(&domainXML, "domain", "", "Domain spec in XML format")
	pflag.Parse()

	logger := log.New(os.Stderr, "replace-roms-sidecar ", log.Ldate)
	if vmiJSON == "" || domainXML == "" {
		logger.Printf("Bad input vmi=%d, domain=%d", len(vmiJSON), len(domainXML))
		os.Exit(1)
	}

	if err := copyFiles(); err != nil {
		logger.Printf("Error while copying OVMF files: %s", err)
		os.Exit(2)
	}

	domainXML, err := onDefineDomain([]byte(vmiJSON), []byte(domainXML))
	if err != nil {
		panic(err)
	}
	fmt.Println(domainXML)
}
