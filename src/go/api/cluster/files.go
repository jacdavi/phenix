package cluster

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"phenix/api/experiment"
	"phenix/util"
	"phenix/util/mm"
	"phenix/util/mm/mmcli"
	"phenix/util/plog"
)

type ImageKind uint8
type CopyStatus func(float64)

const (
	UNKNOWN ImageKind = 1 << iota
	VM_IMAGE
	CONTAINER_IMAGE
	ISO_IMAGE
)

func (k ImageKind) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{"Unknown", "VM", "Container", "ISO"}[k-1])
}

type ImageDetails struct {
	Kind          ImageKind `json:"kind"`
	Name          string    `json:"name"`
	FullPath      string    `json:"fullPath"`
	Size          int       `json:"size"`
	Experiment    *string   `json:"experiment"`
	BackingImages []string  `json:"backingImages"`
	InUse		  bool		`json:"inUse"`
	BackingFor	  string	`json:"backingFor"`
}

var DefaultClusterFiles ClusterFiles = new(MMClusterFiles)
var mmFilesDirectory = util.GetMMFilesDirectory()

type ClusterFiles interface {
	// Get list of VM disk images, container filesystems, or both.
	// Assumes disk images have `.qc2` or `.qcow2` extension.
	// Assumes container filesystems have `_rootfs.tgz` suffix.
	// Alternatively, we could force the use of subdirectories w/ known names
	// (such as `base-images` and `container-fs`).
	GetImages(expName string) ([]ImageDetails, error)
}

func GetImages(expName string) ([]ImageDetails, error) {
	return DefaultClusterFiles.GetImages(expName)
}

type MMClusterFiles struct{}

// Gets images in base directory, plus any images that expName references
// if expName is empty, will check all known experiments
func (MMClusterFiles) GetImages(expName string) ([]ImageDetails, error) {
	// Using a map here to weed out duplicates.
	details := make(map[string]*ImageDetails)

	// Add all the files from the minimega files directory
	if err := getAllFiles(details); err != nil {
		return nil, err
	}

	// Add all files defined in the experiment topology if given; otherwise check all experiments
	if len(expName) > 0 {
		if err := getTopologyFiles(expName, details); err != nil {
			return nil, err
		}
	} else {
		experiments, err := experiment.List()
		if err != nil {
			return nil, err
		}
		for _, exp := range experiments {
			if err := getTopologyFiles(exp.Metadata.Name, details); err != nil {
				return nil, err
			}
		}
	}

	runningVms := mm.GetVMInfo()

	for name := range details {
		if len(details[name].BackingImages) > 0 {
			details[details[name].BackingImages[0]].BackingFor = name
		}
		for _, vm := range runningVms {
			if strings.ReplaceAll(vm.Disk, util.GetMMFilesDirectory(), "") == name {
				details[name].InUse = true
				for _, backing := range details[name].BackingImages {
					details[backing].InUse = true
				}
				break
			}
		}

	}

	var images []ImageDetails

	for name := range details {
		images = append(images, *details[name])
	}


	return images, nil
}

// Get all image files from the minimega files directory
func getAllFiles(details map[string]*ImageDetails) error {

	// First get file listings from mesh, then from headnode.
	commands := []string{"mesh send all file list", "file list"}

	// First, get file listings from cluster nodes.
	cmd := mmcli.NewCommand()

	for _, command := range commands {
		cmd.Command = command

		for _, row := range mmcli.RunTabular(cmd) {

			// Only look in the base directory
			if row["dir"] != "" {
				continue
			}

			baseName := filepath.Base(row["name"])

			// Avoid adding the same image twice
			if _, ok := details[baseName]; ok {
				continue
			}

			image := ImageDetails{
				Name:     baseName,
				FullPath: util.GetMMFullPath(row["name"]),
			}

			if strings.HasSuffix(image.Name, ".qc2") || strings.HasSuffix(image.Name, ".qcow2") {
				image.Kind = VM_IMAGE
				backingImages, err := getBackingImageChain(image)
				if err != nil {
					plog.Warn(fmt.Sprintf("error getting backing images: %v", err))
				} else {
					image.BackingImages = backingImages
				}
			} else if strings.HasSuffix(image.Name, "_rootfs.tgz") {
				image.Kind = CONTAINER_IMAGE
			} else if strings.HasSuffix(image.Name, ".hdd") {
				image.Kind = VM_IMAGE
			} else if strings.HasSuffix(image.Name, ".iso") {
				image.Kind = ISO_IMAGE
			} else {
				continue
			}

			var err error

			image.Size, err = strconv.Atoi(row["size"])
			if err != nil {
				return fmt.Errorf("getting size of file: %w", err)
			}

			details[image.Name] = &image
		}
	}

	return nil

}

// Retrieves all the unique image names defined in the topology
func getTopologyFiles(expName string, details map[string]*ImageDetails) error {
	// Retrieve the experiment
	exp, err := experiment.Get(expName)
	if err != nil {
		return fmt.Errorf("unable to retrieve %v", expName)
	}

	for _, node := range exp.Spec.Topology().Nodes() {
		for _, drive := range node.Hardware().Drives() {
			cmd := mmcli.NewCommand()

			if len(drive.Image()) == 0 {
				continue
			}

			relMMPath, _ := filepath.Rel(mmFilesDirectory, drive.Image())

			if len(relMMPath) == 0 {
				relMMPath = drive.Image()
			}

			cmd.Command = "file list " + relMMPath

			for _, row := range mmcli.RunTabular(cmd) {
				if row["dir"] != "" {
					continue
				}

				baseName := filepath.Base(row["name"])

				// Avoid adding the same image twice
				if _, ok := details[baseName]; ok {
					continue
				}

				image := ImageDetails{
					Name:       baseName,
					FullPath:   util.GetMMFullPath(row["name"]),
					Kind:       VM_IMAGE,
					Experiment: &expName,
				}

				backingImages, err := getBackingImageChain(image)
				if err != nil {
					plog.Warn(fmt.Sprintf("error getting backing images: %v", err))
				} else {
					image.BackingImages = backingImages
				}

				if image.Size, err = strconv.Atoi(row["size"]); err != nil {
					return fmt.Errorf("getting size of file: %w", err)
				}

				details[image.Name] = &image
			}
		}
	}

	return nil
}

func getExperimentNames() (map[string]struct{}, error) {
	experiments, err := experiment.List()
	if err != nil {
		return nil, err
	}

	expNames := make(map[string]struct{})

	for _, exp := range experiments {
		expNames[exp.Spec.ExperimentName()] = struct{}{}
	}

	return expNames, nil
}

func getBackingImageChain(i ImageDetails) ([]string, error) {
	shell := exec.Command("qemu-img", "info", i.FullPath, "--output=json", "--backing-chain")

	res, err := shell.CombinedOutput()
	if err != nil {
		return []string{}, fmt.Errorf("error getting info (%s): %w", string(res), err)
	}

	qemuInfo := []map[string]interface{}{}
	err = json.Unmarshal(res, &qemuInfo)
	if err != nil {
		return []string{}, fmt.Errorf("error  (%s): %w", string(res), err)
	}

	if len(qemuInfo) == 1 {
		return []string{}, nil
	}
	images := []string{}
	for _, q := range qemuInfo {
		backing, ok := q["backing-filename"]
		if !ok {
			break
		}
		images = append(images, backing.(string))
	}
	return images, nil
}
