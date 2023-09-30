package resources

import "syscall"

func MustGetDiskUsage(path string) (free, total, used uint64) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		panic(err)
	}

	// Available blocks * size per block = available space in bytes
	free = stat.Bavail * uint64(stat.Bsize)
	total = stat.Blocks * uint64(stat.Bsize)
	used = total - free

	return
}
