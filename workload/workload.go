package workload

import (
	"context"
	"errors"
	"fmt"
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

	SupportedWorkloads = map[Type]bool{}

	dockerClient *dockercli.Client
)

type (
	Manager struct {
		// pulledImages serves as a much faster cache for checking whether we pulled an image. Is pessimistic (reset on server restart)
		pulledImages syncx.Map[string, bool]
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

func Init() error {
	workloads := strings.Split(utils.SupportedWorkloads, ",")
	if len(workloads) == 0 {
		SupportedWorkloads = map[Type]bool{
			DevWorkload:    true,
			DockerWorkload: true,
		}
	}
	for _, workload := range workloads {
		if string(DevWorkload) != workload && string(DockerWorkload) != workload && string(FirecrackerWorkload) != workload {
			return ErrUnknownWorkloadType
		}
		SupportedWorkloads[Type(workload)] = true
	}

	if SupportedWorkloads[DockerWorkload] {
		cli, err := dockercli.NewClientWithOpts(dockercli.FromEnv)
		if err != nil {
			return fmt.Errorf("error initializing docker client: %w", err)
		}

		dockerClient = cli
	}

	return nil
}

// StartWorkload will attempt to reserve resources and execute. Returns the workload ID.
func (e *Manager) StartWorkload(ctx context.Context, workload Workload) (string, error) {
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

func (e *Manager) StartDockerWorkload(ctx context.Context, workload Workload) error {
	if !SupportedWorkloads[DockerWorkload] {
		return ErrUnsupportedWorkloadType
	}
	payload, ok := workload.Payload.(DockerPayload)
	if !ok {
		return ErrUnknownPayload
	}

	// Pull image (will return fast if already here)
	if _, exists := e.pulledImages.Load(payload.Image); !exists {
		_, err := dockerClient.ImagePull(ctx, payload.Image, dockertypes.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("error in dockerClient.ImagePull: %w", err)
		}
	}

	// TODO: Create container blocking
	// ... the rest in goroutines
	// TODO: Launch timeout goroutine, kill if the wait has not returned first
	// TODO: Docker wait
	return nil
}
