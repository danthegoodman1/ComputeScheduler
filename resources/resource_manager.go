package resources

import (
	"errors"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"sync"
)

type (
	Resources struct {
		// CPU is in millicpu
		CPU int64
		// Memory is in MiB
		Memory int64
		// // PersistentDisk is in bytes
		// PersistentDisk int64
	}

	ResourceManager struct {
		TotalResources Resources
		freeResources  Resources
		allocations    map[string]Resources
		mu             *sync.Mutex
	}
)

const (
	MinCPUAllocation = 100
	MinMemAllocation = 128
)

var (
	Manager ResourceManager

	ErrNotEnoughCPU    = errors.New("not enough cpu available")
	ErrNotEnoughMemory = errors.New("not enough memory available")

	ErrCPUAllocationTooLow = errors.New("cpu allocation too low")
	ErrMemAllocationTooLow = errors.New("mem allocation too low")

	ErrAllocationNotFound = errors.New("allocation not found")
)

func InitResourceManager() {
	Manager = ResourceManager{
		TotalResources: Resources{
			CPU:    utils.ReservedCPU,
			Memory: utils.ReservedMem,
		},
		freeResources: Resources{
			CPU:    utils.ReservedCPU,
			Memory: utils.ReservedMem,
		},
		allocations: map[string]Resources{},
		mu:          &sync.Mutex{},
	}
}

func (manager *ResourceManager) AllocateResources(resources Resources) (string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Verify we have the resources
	if manager.freeResources.CPU < resources.CPU {
		return "", ErrNotEnoughCPU
	}
	if manager.freeResources.Memory < resources.Memory {
		return "", ErrNotEnoughMemory
	}

	if resources.CPU < MinCPUAllocation {
		return "", ErrCPUAllocationTooLow
	}
	if resources.Memory < MinMemAllocation {
		return "", ErrMemAllocationTooLow
	}

	// Reserve the resources
	allocID := utils.GenRandomID("alloc_")
	manager.freeResources.CPU -= resources.CPU
	manager.freeResources.Memory -= resources.Memory
	manager.allocations[allocID] = resources

	// TODO: Metrics free resources

	return allocID, nil
}

func (manager *ResourceManager) FreeAllocation(allocID string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	resources, exists := manager.allocations[allocID]
	if !exists {
		return ErrAllocationNotFound
	}

	manager.freeResources.CPU += resources.CPU
	manager.freeResources.Memory += resources.Memory
	delete(manager.allocations, allocID)

	// TODO: Metrics free resources
	return nil
}

func (manager *ResourceManager) GetFreeResources() Resources {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.freeResources
}
