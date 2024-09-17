package cluster

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"phenix/api/experiment"
	"phenix/util"
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
	Size          string     `json:"size"`
	VirtualSize   string     `json:"virtualSize"`
	Experiment    *string   `json:"experiment"`
	BackingImages []string  `json:"backingImages"`
	InUse         bool      `json:"inUse"`
}

var (
	DefaultClusterFiles ClusterFiles = new(MMClusterFiles)
	mmFilesDirectory = util.GetMMFilesDirectory()
	knownImageExtensions = []string{".qcow2", ".qc2", "_rootfs.tgz", ".hdd", ".iso"}
)

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
	details := make(map[string]ImageDetails)

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

	plog.Info("GOT", "images", details)
	var images []ImageDetails
	for name := range details {
		images = append(images, details[name])
	}

	return images, nil
}

// Get all image files from the minimega files directory
func getAllFiles(details map[string]ImageDetails) error {

	// First get file listings from mesh, then from headnode.
	commands := []string{"mesh send all file list", "file list"}

	// First, get file listings from cluster nodes.
	cmd := mmcli.NewCommand()

	for _, command := range commands {
		cmd.Command = command

		for _, row := range mmcli.RunTabular(cmd) {
			plog.Info("FILE", "file", row["name"])
			if _, ok := details[row["name"]]; row["dir"] == "" && !ok {
				for _, image := range resolveImage(filepath.Join(mmFilesDirectory, row["name"])) {
					if _, ok := details[image.Name]; !ok {
						details[image.Name] = image
					}
				}
			}
		}
	}

	return nil

}

// Retrieves all the unique image names defined in the topology
func getTopologyFiles(expName string, details map[string]ImageDetails) error {
	// Retrieve the experiment
	exp, err := experiment.Get(expName)
	if err != nil {
		return fmt.Errorf("unable to retrieve %v", expName)
	}

	for _, node := range exp.Spec.Topology().Nodes() {
		for _, drive := range node.Hardware().Drives() {
			if len(drive.Image()) == 0 {
				continue
			}

			relMMPath, _ := filepath.Rel(mmFilesDirectory, drive.Image())
			if _, ok := details[filepath.Base(relMMPath)]; !ok {
				for _, image := range resolveImage(relMMPath) {
					if _, ok := details[image.Name]; !ok {
						details[image.Name] = image
					}
				}
			}
		}
	}

	return nil
}

func resolveImage(path string) ([]ImageDetails) {
	imageDetails := []ImageDetails{}


	knownFormat := false
	for _, f := range knownImageExtensions {
		if filepath.Ext(path) == f {
			knownFormat = true
			break
		}
	}
	if !knownFormat {
		plog.Debug("File didn't match know image extensions: %s", path)
		return imageDetails
	}


	cmd := mmcli.NewCommand()
	cmd.Command = fmt.Sprintf("disk info %v recursive", path)
	plog.Info("CMD", "cmd", cmd.Command)
	images := mmcli.RunTabular(cmd)
	plog.Info("FILE", "images", images)

	for i, row := range images {
		image := ImageDetails{
			Name: filepath.Base(row["image"]),
			FullPath: row["image"],
			Size: row["disksize"],
			VirtualSize: row["virtualsize"],
		}

		if row["format"] == "qcow2" {
			image.Kind = VM_IMAGE
		} else if strings.HasSuffix(image.Name, "_rootfs.tgz") {
			image.Kind = CONTAINER_IMAGE
		} else if strings.HasSuffix(image.Name, ".hdd") {
			image.Kind = VM_IMAGE
		} else if strings.HasSuffix(image.Name, ".iso") {
			image.Kind = ISO_IMAGE
		} 

		var err error
		image.InUse, err = strconv.ParseBool(row["inuse"])
		if err != nil {
			plog.Warn("could not determine if image in use", "image", path)
		}

		backingChain := []string{}
		for _, backing := range images[i:] {
			backingChain = append(backingChain, filepath.Base(backing["image"]))
		}
		image.BackingImages = backingChain
		imageDetails = append(imageDetails, image)
	}

	return imageDetails
}
