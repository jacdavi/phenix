package disk

var DefaultDiskFiles DiskFiles = new(MMDiskFiles)

func GetImages(expName string) ([]Details, error) {
	return DefaultDiskFiles.GetImages(expName)
}

func CommitDisk(path string) error {
	return DefaultDiskFiles.CommitDisk(path)
}

func SnapshotDisk(src, dst string) error {
	return DefaultDiskFiles.SnapshotDisk(src, dst)
}