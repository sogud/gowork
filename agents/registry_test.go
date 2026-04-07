package agents

import (
	"sync"
	"testing"
)

// mockAgent implements Agent interface for testing
type mockAgent struct {
	name        string
	description string
}

func (m *mockAgent) Name() string        { return m.name }
func (m *mockAgent) Description() string { return m.description }

func TestRegistry_Register(t *testing.T) {
	t.Run("successfully register an agent", func(t *testing.T) {
		r := NewRegistry()
		agent := &mockAgent{name: "test-agent", description: "test description"}

		err := r.Register(agent)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify agent was registered
		retrieved, err := r.Get("test-agent")
		if err != nil {
			t.Errorf("expected to retrieve agent, got error: %v", err)
		}
		if retrieved.Name() != "test-agent" {
			t.Errorf("expected agent name 'test-agent', got %s", retrieved.Name())
		}
	})

	t.Run("prevent duplicate registration", func(t *testing.T) {
		r := NewRegistry()
		agent1 := &mockAgent{name: "duplicate", description: "first"}
		agent2 := &mockAgent{name: "duplicate", description: "second"}

		err := r.Register(agent1)
		if err != nil {
			t.Errorf("expected no error on first registration, got %v", err)
		}

		err = r.Register(agent2)
		if err == nil {
			t.Error("expected error on duplicate registration, got nil")
		}
	})

	t.Run("reject nil agent", func(t *testing.T) {
		r := NewRegistry()
		err := r.Register(nil)
		if err == nil {
			t.Error("expected error when registering nil agent")
		}
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("get existing agent", func(t *testing.T) {
		r := NewRegistry()
		agent := &mockAgent{name: "existing", description: "test"}
		_ = r.Register(agent)

		retrieved, err := r.Get("existing")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if retrieved.Name() != "existing" {
			t.Errorf("expected 'existing', got %s", retrieved.Name())
		}
	})

	t.Run("get non-existent agent", func(t *testing.T) {
		r := NewRegistry()

		_, err := r.Get("non-existent")
		if err == nil {
			t.Error("expected error for non-existent agent")
		}
	})
}

func TestRegistry_GetAll(t *testing.T) {
	t.Run("get all agents returns copy", func(t *testing.T) {
		r := NewRegistry()
		agent1 := &mockAgent{name: "agent1", description: "first"}
		agent2 := &mockAgent{name: "agent2", description: "second"}

		_ = r.Register(agent1)
		_ = r.Register(agent2)

		all := r.GetAll()
		if len(all) != 2 {
			t.Errorf("expected 2 agents, got %d", len(all))
		}

		// Modify returned slice should not affect registry
		all["agent3"] = &mockAgent{name: "agent3", description: "third"}

		allAgain := r.GetAll()
		if len(allAgain) != 2 {
			t.Error("GetAll should return a copy, not internal map")
		}
	})

	t.Run("get all from empty registry", func(t *testing.T) {
		r := NewRegistry()
		all := r.GetAll()
		if all == nil {
			t.Error("GetAll should return empty map, not nil")
		}
		if len(all) != 0 {
			t.Errorf("expected empty map, got %d items", len(all))
		}
	})
}

func TestRegistry_List(t *testing.T) {
	t.Run("list all agent names", func(t *testing.T) {
		r := NewRegistry()
		agent1 := &mockAgent{name: "alpha", description: "first"}
		agent2 := &mockAgent{name: "beta", description: "second"}
		agent3 := &mockAgent{name: "gamma", description: "third"}

		_ = r.Register(agent1)
		_ = r.Register(agent2)
		_ = r.Register(agent3)

		names := r.List()
		if len(names) != 3 {
			t.Errorf("expected 3 names, got %d", len(names))
		}

		// Check all names are present
		nameSet := make(map[string]bool)
		for _, name := range names {
			nameSet[name] = true
		}

		expected := []string{"alpha", "beta", "gamma"}
		for _, e := range expected {
			if !nameSet[e] {
				t.Errorf("expected name %s in list", e)
			}
		}
	})

	t.Run("list from empty registry", func(t *testing.T) {
		r := NewRegistry()
		names := r.List()
		if len(names) != 0 {
			t.Errorf("expected empty list, got %d items", len(names))
		}
	})
}

func TestRegistry_Concurrency(t *testing.T) {
	t.Run("concurrent registrations", func(t *testing.T) {
		r := NewRegistry()
		var wg sync.WaitGroup

		// Try to register 100 agents concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				agent := &mockAgent{
					name:        string(rune(id)),
					description: "concurrent test",
				}
				_ = r.Register(agent)
			}(i)
		}

		wg.Wait()

		// All registrations should be safe
		all := r.GetAll()
		if len(all) == 0 {
			t.Error("expected at least one successful registration")
		}
	})

	t.Run("concurrent read and write", func(t *testing.T) {
		r := NewRegistry()
		var wg sync.WaitGroup

		// Writer goroutines
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				agent := &mockAgent{
					name:        string(rune(id)),
					description: "writer",
				}
				_ = r.Register(agent)
			}(i)
		}

		// Reader goroutines
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = r.GetAll()
				_ = r.List()
			}()
		}

		wg.Wait()
	})
}