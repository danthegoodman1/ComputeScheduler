package workload

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/gologger"
	"github.com/danthegoodman1/GoAPITemplate/resources"
	"github.com/danthegoodman1/GoAPITemplate/syncx"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	dockertypes "github.com/docker/docker/api/types"
	dockercli "github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"strings"
	"time"
)

var (
	ErrUnknownWorkloadType     = errors.New("unknown workload type")
	ErrUnknownPayload          = errors.New("unknown payload")
	ErrUnsupportedWorkloadType = errors.New("unsupported workload type")

	dockerClient *dockercli.Client
)

type (
	WorkloadManager struct {
		// pulledImages serves as a much faster cache for checking whether we pulled an image. Is pessimistic (reset on server restart)
		pulledImages       syncx.Map[string, bool]
		SupportedWorkloads map[Type]bool
	}
	Workload struct {
		Resources resources.Resources
		Type      Type
		Timeout   *time.Duration
		Payload   any
	}

	Type string
)

const (
	DevWorkload         Type = "dev"
	DockerWorkload      Type = "docker"
	FirecrackerWorkload Type = "firecracker"
)

var (
	logger  = gologger.NewLogger()
	Manager *WorkloadManager
)

func Init() error {
	Manager = &WorkloadManager{
		pulledImages: syncx.NewMap[string, bool](),
	}

	// Add in preinstalled image
	for _, img := range strings.Split(utils.PreInstalledImages, ",") {
		if img == "" {
			continue
		}
		img := strings.TrimSpace(img)
		logger.Debug().Msgf("Adding preinstalled image: %s", img)
		Manager.pulledImages.Load(img)
	}

	if utils.SupportedWorkloads == "" {
		logger.Debug().Msg("using default workloads of docker and dev")
		Manager.SupportedWorkloads = map[Type]bool{
			DevWorkload:    true,
			DockerWorkload: true,
		}
	} else {
		workloads := strings.Split(utils.SupportedWorkloads, ",")
		for _, workload := range workloads {
			if workload == "" {
				continue
			}
			if string(DevWorkload) != workload && string(DockerWorkload) != workload && string(FirecrackerWorkload) != workload {
				return ErrUnknownWorkloadType
			}
			Manager.SupportedWorkloads[Type(workload)] = true
		}
	}

	if Manager.SupportedWorkloads[DockerWorkload] {
		cli, err := dockercli.NewClientWithOpts(dockercli.FromEnv)
		if err != nil {
			return fmt.Errorf("error initializing docker client: %w", err)
		}

		dockerClient = cli
	}

	return nil
}

// StartWorkload will attempt to reserve resources and execute. Returns the workload ID.
func (e *WorkloadManager) StartWorkload(ctx context.Context, workload Workload) (string, error) {
	workloadID := utils.GenRandomID("wrkl_")
	logger := zerolog.Ctx(ctx).With().Str("workloadID", workloadID).Interface("workload", workload).Logger()
	logger.Debug().Msg("starting workload")

	// Launch the compute job
	switch workload.Type {
	case DevWorkload:
		break
	case DockerWorkload:
		err := e.StartDockerWorkload(ctx, workload)
		if err != nil {
			return "", fmt.Errorf("error in StartDockerWorkload: %w", err)
		}
	default:
		return "", ErrUnknownWorkloadType
	}

	return workloadID, nil
}

type (
	DockerPayload struct {
		Image string
	}
)

func (e *WorkloadManager) StartDockerWorkload(ctx context.Context, workload Workload) error {
	if !e.SupportedWorkloads[DockerWorkload] {
		return ErrUnsupportedWorkloadType
	}
	payload, ok := workload.Payload.(DockerPayload)
	if !ok {
		return ErrUnknownPayload
	}

	// Pull image if we don't (potentially) have it
	if _, exists := e.pulledImages.Load(payload.Image); !exists {
		_, err := dockerClient.ImagePull(ctx, payload.Image, dockertypes.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("error in dockerClient.ImagePull: %w", err)
		}
		e.pulledImages.Store(payload.Image, true)
	}

	// TODO: Create container blocking
	// ... the rest in goroutines
	// TODO: Launch timeout goroutine, kill if the wait has not returned first
	// TODO: Docker wait
	return nil
}
