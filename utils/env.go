package utils

import (
	"github.com/samber/lo"
	"os"
)

const (
	WorkerRole    = "WORKER"
	SchedulerRole = "SCHEDULER"
)

var (
	Env      = os.Getenv("ENV")
	Hostname = os.Getenv("HOSTNAME")

	// Role is WorkerRole by default
	Role        = lo.Ternary(os.Getenv("ROLE") == "", WorkerRole, os.Getenv("ROLE"))
	IsWorker    = Role == WorkerRole
	IsScheduler = Role == SchedulerRole

	ReservedCPU = MustGetEnvInt("RESERVED_CPU")
	ReservedMem = MustGetEnvInt("RESERVED_MEM")

	// dev,docker,firecracker
	SupportedWorkloads = os.Getenv("SUPPORTED_WORKLOADS")

	// CSV of images that are already avialable, won't be pulled - image1,image2,image3:latest,...
	PreInstalledImages = os.Getenv("PREINSTALLED_IMAGES")
)
