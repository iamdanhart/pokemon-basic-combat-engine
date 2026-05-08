package engine

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestVM(t *testing.T, script string) *VM {
	t.Helper()
	vm := NewVM()
	if err := vm.L.DoString(script); err != nil {
		t.Fatalf("lua setup failed: %v", err)
	}
	return vm
}

// --- loadSpecies ---

func TestLoadSpecies_singleType(t *testing.T) {
	vm := newTestVM(t, `
		Pokemon = {
			charmander = { name="Charmander", types={"fire"}, hp=39, atk=52, def=43, spd=65 },
		}
	`)
	if err := vm.loadSpecies(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, ok := vm.Species["charmander"]
	if !ok {
		t.Fatal("charmander not found in Species map")
	}
	if s.Name != "Charmander" {
		t.Errorf("Name: got %q, want %q", s.Name, "Charmander")
	}
	if s.NumTypes != 1 {
		t.Errorf("NumTypes: got %d, want 1", s.NumTypes)
	}
	if s.Types[0] != TypeFire {
		t.Errorf("Types[0]: got %v, want Fire", s.Types[0])
	}
	if s.BaseHP != 39 {
		t.Errorf("BaseHP: got %d, want 39", s.BaseHP)
	}
	if s.BaseAtk != 52 {
		t.Errorf("BaseAtk: got %d, want 52", s.BaseAtk)
	}
	if s.BaseDef != 43 {
		t.Errorf("BaseDef: got %d, want 43", s.BaseDef)
	}
	if s.BaseSpd != 65 {
		t.Errorf("BaseSpd: got %d, want 65", s.BaseSpd)
	}
}

func TestLoadSpecies_dualType(t *testing.T) {
	vm := newTestVM(t, `
		Pokemon = {
			bulbasaur = { name="Bulbasaur", types={"grass","poison"}, hp=45, atk=49, def=49, spd=45 },
		}
	`)
	if err := vm.loadSpecies(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := vm.Species["bulbasaur"]
	if s.NumTypes != 2 {
		t.Errorf("NumTypes: got %d, want 2", s.NumTypes)
	}
	if s.Types[0] != TypeGrass {
		t.Errorf("Types[0]: got %v, want Grass", s.Types[0])
	}
	if s.Types[1] != TypePoison {
		t.Errorf("Types[1]: got %v, want Poison", s.Types[1])
	}
}

func TestLoadSpecies_unknownType(t *testing.T) {
	vm := newTestVM(t, `
		Pokemon = {
			fake = { name="Fake", types={"banana"}, hp=50, atk=50, def=50, spd=50 },
		}
	`)
	if err := vm.loadSpecies(); err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestLoadSpecies_missingPokemonTable(t *testing.T) {
	vm := NewVM()
	if err := vm.loadSpecies(); err == nil {
		t.Error("expected error when Pokemon global is missing, got nil")
	}
}

// --- loadMoves ---

func TestLoadMoves_damagingMove(t *testing.T) {
	vm := newTestVM(t, `
		Moves = {
			ember = { name="Ember", type="fire", category="special", power=40, accuracy=100, pp=25 },
		}
	`)
	if err := vm.loadMoves(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := vm.Moves["ember"]
	if !ok {
		t.Fatal("ember not found in Moves map")
	}
	if m.Name != "Ember" {
		t.Errorf("Name: got %q, want %q", m.Name, "Ember")
	}
	if m.Type != TypeFire {
		t.Errorf("Type: got %v, want Fire", m.Type)
	}
	if m.Category != CategorySpecial {
		t.Errorf("Category: got %v, want Special", m.Category)
	}
	if m.Power != 40 {
		t.Errorf("Power: got %d, want 40", m.Power)
	}
	if m.Accuracy != 100 {
		t.Errorf("Accuracy: got %d, want 100", m.Accuracy)
	}
	if m.PP != 25 || m.PPMax != 25 {
		t.Errorf("PP: got %d/%d, want 25/25", m.PP, m.PPMax)
	}
	if m.EffectFunc != "" {
		t.Errorf("EffectFunc: expected empty, got %q", m.EffectFunc)
	}
}

func TestLoadMoves_statusMoveWithEffect(t *testing.T) {
	vm := newTestVM(t, `
		Moves = {
			thunder_wave = {
				name="Thunder Wave", type="electric", category="status",
				power=0, accuracy=90, pp=20,
				effect = function(actor, target)
					if target.status == 0 then target.status = 2 end
				end,
			},
		}
	`)
	if err := vm.loadMoves(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := vm.Moves["thunder_wave"]
	if m.EffectFunc == "" {
		t.Error("expected EffectFunc to be set for status move with effect")
	}
	if vm.L.GetGlobal(m.EffectFunc) == lua.LNil {
		t.Errorf("expected Lua global %q to be registered", m.EffectFunc)
	}
}

func TestLoadMoves_unknownCategory(t *testing.T) {
	vm := newTestVM(t, `
		Moves = {
			bad = { name="Bad", type="normal", category="banana", power=0, accuracy=100, pp=10 },
		}
	`)
	if err := vm.loadMoves(); err == nil {
		t.Error("expected error for unknown category, got nil")
	}
}

func TestLoadMoves_missingMovesTable(t *testing.T) {
	vm := NewVM()
	if err := vm.loadMoves(); err == nil {
		t.Error("expected error when Moves global is missing, got nil")
	}
}