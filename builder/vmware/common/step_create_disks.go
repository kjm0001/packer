package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates the virtual disks for the VM.
//
// Uses:
//   config *config
//   driver Driver
//   ui     packer.Ui
//
// Produces:
//   disk_full_paths ([]string) - The full paths to all created disks
type StepCreateDisks struct {
	OutputDir          *string
	CreateMainDisk     bool
	DiskName           string
	MainDiskSize       uint
	AdditionalDiskSize []uint
	DiskAdapterType    string
	DiskTypeId         string
}

func (s *StepCreateDisks) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating required virtual machine disks")

	// Users can configure disks at several locations in the template so
	// first collate all the disk requirements
	var diskFullPaths, diskSizes []string
	// The 'main' or 'default' disk, only used in vmware-iso
	if s.CreateMainDisk {
		log.Printf("Megan creating main disk.")
		diskFullPaths = append(diskFullPaths, filepath.Join(*s.OutputDir, s.DiskName+".vmdk"))
		diskSizes = append(diskSizes, fmt.Sprintf("%dM", uint64(s.MainDiskSize)))
	}
	// Additional disks
	if len(s.AdditionalDiskSize) > 0 {
		for i, diskSize := range s.AdditionalDiskSize {
			path := filepath.Join(*s.OutputDir, fmt.Sprintf("%s-%d.vmdk", s.DiskName, i+1))
			diskFullPaths = append(diskFullPaths, path)
			size := fmt.Sprintf("%dM", uint64(diskSize))
			diskSizes = append(diskSizes, size)
		}
	}

	// Create all required disks
	for i, diskFullPath := range diskFullPaths {
		log.Printf("[INFO] Creating disk with Path: %s and Size: %s", diskFullPath, diskSizes[i])
		// Additional disks currently use the same adapter type and disk
		// type as specified for the main disk
		if err := driver.CreateDisk(diskFullPath, diskSizes[i], s.DiskAdapterType, s.DiskTypeId); err != nil {
			err := fmt.Errorf("Error creating disk: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Stash the disk paths so we can retrieve later e.g. when compacting
	state.Put("disk_full_paths", diskFullPaths)
	return multistep.ActionContinue
}

func (s *StepCreateDisks) Cleanup(multistep.StateBag) {}
