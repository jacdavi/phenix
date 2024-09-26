package disk

var DefaultDiskFiles DiskFiles = new(MMDiskFiles)

func GetImages(expName string) ([]Details, error) {
	return DefaultDiskFiles.GetImages(expName)
}

func GetImage(path string) (Details, error) {
	return DefaultDiskFiles.GetImage(path)
}

func CommitDisk(path string) error {
	return DefaultDiskFiles.CommitDisk(path)
}

func SnapshotDisk(src, dst string) error {
	return DefaultDiskFiles.SnapshotDisk(src, dst)
}

func RebaseDisk(src, dst string, unsafe bool) error {
	return DefaultDiskFiles.RebaseDisk(src, dst, unsafe)
}

func CloneDisk(src, dst string) error {
	return DefaultDiskFiles.CloneDisk(src, dst)
}

func RenameDisk(src, dst string) error {
	return DefaultDiskFiles.RenameDisk(src, dst)
}

func DeleteDisk(src string) error {
	return DefaultDiskFiles.DeleteDisk(src)
}