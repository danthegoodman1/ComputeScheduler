package resources

import (
	"errors"
	"sync"
	"testing"
)

func TestResourceManager(t *testing.T) {
	manager := ResourceManager{
		TotalResources: Resources{
			CPU:    1000,
			Memory: 1280,
		},
		freeResources: Resources{
			CPU:    1000,
			Memory: 1280,
		},
		allocations: map[string]Resources{},
		mu:          &sync.Mutex{},
	}

	t.Log("test allocate")
	allocID, err := manager.AllocateResources(Resources{
		CPU:    100,
		Memory: 128,
	})
	if err != nil {
		t.Fatal(err)
	}
	resources := manager.GetFreeResources()
	if resources.Memory != manager.TotalResources.Memory-128 || resources.CPU != manager.TotalResources.CPU-100 {
		t.Fatalf("mismatch of resources manager: %+v\n\nresources:%+v", manager, resources)
	}

	t.Log("test free")
	err = manager.FreeAllocation(allocID)
	resources = manager.GetFreeResources()
	if resources.Memory != manager.TotalResources.Memory || resources.CPU != manager.TotalResources.CPU {
		t.Fatalf("mismatch of resources manager: %+v\n\nresources:%+v", manager, resources)
	}

	t.Log("test allocate too much")
	for i := 0; i < 9; i++ {
		allocID, err = manager.AllocateResources(Resources{
			CPU:    100,
			Memory: 128,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	allocID, err = manager.AllocateResources(Resources{
		CPU:    101,
		Memory: 128,
	})
	if !errors.Is(err, ErrNotEnoughCPU) {
		t.Fatal("failed not enough cpu check", err)
	}
	allocID, err = manager.AllocateResources(Resources{
		CPU:    100,
		Memory: 129,
	})
	if !errors.Is(err, ErrNotEnoughMemory) {
		t.Fatal("failed not enough mem check", err)
	}

	t.Log("test allocate too little")
	allocID, err = manager.AllocateResources(Resources{
		CPU:    99,
		Memory: 128,
	})
	if !errors.Is(err, ErrCPUAllocationTooLow) {
		t.Fatal("failed too little cpu check", err)
	}
	allocID, err = manager.AllocateResources(Resources{
		CPU:    100,
		Memory: 127,
	})
	if !errors.Is(err, ErrMemAllocationTooLow) {
		t.Fatal("failed too little mem check", err)
	}

	t.Log("test free nothing")
	err = manager.FreeAllocation("blah")
	if !errors.Is(err, ErrAllocationNotFound) {
		t.Fatal("failed allocation not found check", err)
	}
}
