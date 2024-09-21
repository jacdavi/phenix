package disk

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

type Kind uint8
type CopyStatus func(float64)

const (
	UNKNOWN Kind = 1 << iota
	VM_IMAGE
	CONTAINER_IMAGE
	ISO_IMAGE
)

func (k Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k Kind) String() string {
	switch k {
	case VM_IMAGE:
		return "VM"
	case CONTAINER_IMAGE:
		return "Container"
	case ISO_IMAGE:
		return "ISO"
	default:
		return "Unknown"

	}
}

func StringToKind(kind string) Kind {
	switch strings.ToLower(kind) {
	case "vm":
		return VM_IMAGE
	case "iso":
		return ISO_IMAGE
	case "container":
		return CONTAINER_IMAGE
	default:
		return UNKNOWN
	}
}

type Details struct {
	Kind          Kind `json:"kind"`
	Name          string    `json:"name"`
	FullPath      string    `json:"fullPath"`
	Size          string    `json:"size"`
	VirtualSize   string    `json:"virtualSize"`
	Experiment    *string   `json:"experiment"`
	BackingImages []string  `json:"backingImages"`
	InUse         bool      `json:"inUse"`
}

var (
	DefaultDiskFiles  DiskFiles = new(MMDiskFiles)
	mmFilesDirectory                  = util.GetMMFilesDirectory()
	knownImageExtensions              = []string{".qcow2", ".qc2", "_rootfs.tgz", ".hdd", ".iso"}
)

type DiskFiles interface {
	// Get list of VM disk images, container filesystems, or both.
	// Assumes disk images have `.qc2` or `.qcow2` extension.
	// Assumes container filesystems have `_rootfs.tgz` suffix.
	// Alternatively, we could force the use of subdirectories w/ known names
	// (such as `base-images` and `container-fs`).
	GetImages(expName string) ([]Details, error)

	CommitDisk(path string) error

	SnapshotDisk(src, dst string) error
}

func GetImages(expName string) ([]Details, error) {
	return DefaultDiskFiles.GetImages(expName)
}

func CommitDisk(path string) error {
	return DefaultDiskFiles.CommitDisk(path)
}

func SnapshotDisk(src, dst string) error {
	return DefaultDiskFiles.SnapshotDisk(src, dst)
}

type MMDiskFiles struct{}


func (MMDiskFiles) CommitDisk(path string) error {
	cmd := mmcli.NewCommand()
	cmd.Command = fmt.Sprintf("disk commit %s", path)
	_, err := mmcli.SingleDataResponse(mmcli.Run(cmd))
	return err
}

func (MMDiskFiles) SnapshotDisk(src, dst string) error {
	cmd := mmcli.NewCommand()
	cmd.Command = fmt.Sprintf("disk snapshot %s %s", src, dst)
	_, err := mmcli.SingleDataResponse(mmcli.Run(cmd))
	return err
}

// Gets images in base directory, plus any images that expName references
// if expName is empty, will check all known experiments
func (MMDiskFiles) GetImages(expName string) ([]Details, error) {
	// Using a map here to weed out duplicates.
	details := make(map[string]Details)

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

	var images []Details
	for name := range details {
		images = append(images, details[name])
	}

	return images, nil
}

// Get all image files from the minimega files directory
func getAllFiles(details map[string]Details) error {

	// First get file listings from mesh, then from headnode.
	commands := []string{"mesh send all file list", "file list"}

	// First, get file listings from cluster nodes.
	cmd := mmcli.NewCommand()

	for _, command := range commands {
		cmd.Command = command

		for _, row := range mmcli.RunTabular(cmd) {
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
func getTopologyFiles(expName string, details map[string]Details) error {
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
			path := drive.Image()
			if !filepath.IsAbs(path) {
				path = filepath.Join(mmFilesDirectory, path)
			}
			if _, ok := details[filepath.Base(path)]; !ok {
				for _, image := range resolveImage(path) {
					if _, ok := details[image.Name]; !ok {
						details[image.Name] = image
					}
				}
			}
		}
	}

	return nil
}

func resolveImage(path string) []Details {
	imageDetails := []Details{}

	knownFormat := false
	for _, f := range knownImageExtensions {
		if strings.HasSuffix(path, f) {
			knownFormat = true
			break
		}
	}
	if !knownFormat {
		plog.Debug("file didn't match know image extensions: %s", "path", path)
		return imageDetails
	}

	cmd := mmcli.NewCommand()
	cmd.Command = fmt.Sprintf("disk info %v recursive", path)
	images := mmcli.RunTabular(cmd)

	for i, row := range images {
		image := Details{
			Name:        filepath.Base(row["image"]),
			FullPath:    row["image"],
			Size:        row["disksize"],
			VirtualSize: row["virtualsize"],
		}

		if row["format"] == "qcow2" {
			image.Kind = VM_IMAGE
			backingChain := []string{}
			for _, backing := range images[i+1:] {
				backingChain = append(backingChain, filepath.Base(backing["image"]))
			}
			image.BackingImages = backingChain
		} else if strings.HasSuffix(image.Name, "_rootfs.tgz") {
			image.Kind = CONTAINER_IMAGE
		} else if strings.HasSuffix(image.Name, ".hdd") {
			image.Kind = VM_IMAGE
		} else if strings.HasSuffix(image.Name, ".iso") {
			image.Kind = ISO_IMAGE
		} else {
			image.Kind = UNKNOWN
		}

		var err error
		image.InUse, err = strconv.ParseBool(row["inuse"])
		if err != nil {
			plog.Warn("could not determine if image in use", "image", path)
		}

		imageDetails = append(imageDetails, image)
	}

	return imageDetails
}
