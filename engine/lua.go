package engine

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type StatusHooks struct {
	OnTurnStart string // Lua global name, empty if no hook
	OnTurnEnd   string
}

type VM struct {
	L           *lua.LState
	Species     map[string]*Species
	Moves       map[string]*Move
	StatusHooks map[StatusEffect]StatusHooks
}

func NewVM() *VM {
	L := lua.NewState()
	if err := L.DoString(fmt.Sprintf("math.randomseed(%d)", time.Now().UnixNano())); err != nil {
		panic(fmt.Sprintf("failed to seed Lua RNG: %v", err))
	}

	// Expose StatusEffect constants to Lua so scripts use Status.paralysis rather than magic numbers.
	statusConsts := L.NewTable()
	L.SetField(statusConsts, "none",      lua.LNumber(StatusNone))
	L.SetField(statusConsts, "burn",      lua.LNumber(StatusBurn))
	L.SetField(statusConsts, "paralysis", lua.LNumber(StatusParalyze))
	L.SetField(statusConsts, "sleep",     lua.LNumber(StatusSleep))
	L.SetField(statusConsts, "poison",    lua.LNumber(StatusPoison))
	L.SetField(statusConsts, "freeze",     lua.LNumber(StatusFreeze))
	L.SetField(statusConsts, "bad_poison", lua.LNumber(StatusBadPoison))
	L.SetGlobal("Status", statusConsts)

	return &VM{
		L:           L,
		Species:     make(map[string]*Species),
		Moves:       make(map[string]*Move),
		StatusHooks: make(map[StatusEffect]StatusHooks),
	}
}

func (vm *VM) Close() {
	vm.L.Close()
}

func (vm *VM) LoadScripts(dir string) error {
	if err := vm.L.DoFile(dir + "/pokemon.lua"); err != nil {
		return fmt.Errorf("loading pokemon.lua: %w", err)
	}
	if err := vm.loadSpecies(); err != nil {
		return fmt.Errorf("parsing pokemon.lua: %w", err)
	}

	if err := vm.L.DoFile(dir + "/moves.lua"); err != nil {
		return fmt.Errorf("loading moves.lua: %w", err)
	}
	if err := vm.loadMoves(); err != nil {
		return fmt.Errorf("parsing moves.lua: %w", err)
	}

	if err := vm.L.DoFile(dir + "/status_effects.lua"); err != nil {
		return fmt.Errorf("loading status_effects.lua: %w", err)
	}
	if err := vm.loadStatusHooks(); err != nil {
		return fmt.Errorf("parsing status_effects.lua: %w", err)
	}

	return nil
}

func (vm *VM) loadSpecies() error {
	tbl, ok := vm.L.GetGlobal("Pokemon").(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected Pokemon to be a table")
	}

	var outerErr error
	tbl.ForEach(func(key, val lua.LValue) {
		if outerErr != nil {
			return
		}
		entry, ok := val.(*lua.LTable)
		if !ok {
			return
		}

		s := &Species{}
		s.Name = lua.LVAsString(entry.RawGetString("name"))

		typesTbl, ok := entry.RawGetString("types").(*lua.LTable)
		if !ok {
			outerErr = fmt.Errorf("species %q: types must be a table", s.Name)
			return
		}
		typesTbl.ForEach(func(_, tv lua.LValue) {
			if outerErr != nil || s.NumTypes >= 2 {
				return
			}
			t, err := TypeFromString(lua.LVAsString(tv))
			if err != nil {
				outerErr = fmt.Errorf("species %q: %w", s.Name, err)
				return
			}
			s.Types[s.NumTypes] = t
			s.NumTypes++
		})
		if outerErr != nil {
			return
		}

		s.BaseHP = int(lua.LVAsNumber(entry.RawGetString("hp")))
		s.BaseAtk = int(lua.LVAsNumber(entry.RawGetString("atk")))
		s.BaseDef = int(lua.LVAsNumber(entry.RawGetString("def")))
		s.BaseSpd = int(lua.LVAsNumber(entry.RawGetString("spd")))

		vm.Species[lua.LVAsString(key)] = s
	})

	return outerErr
}

func (vm *VM) loadMoves() error {
	tbl, ok := vm.L.GetGlobal("Moves").(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected Moves to be a table")
	}

	var outerErr error
	tbl.ForEach(func(key, val lua.LValue) {
		if outerErr != nil {
			return
		}
		entry, ok := val.(*lua.LTable)
		if !ok {
			return
		}

		m := &Move{}
		m.Name = lua.LVAsString(entry.RawGetString("name"))

		t, err := TypeFromString(lua.LVAsString(entry.RawGetString("type")))
		if err != nil {
			outerErr = fmt.Errorf("move %q: %w", m.Name, err)
			return
		}
		m.Type = t

		switch lua.LVAsString(entry.RawGetString("category")) {
		case "physical":
			m.Category = CategoryPhysical
		case "special":
			m.Category = CategorySpecial
		case "status":
			m.Category = CategoryStatus
		default:
			outerErr = fmt.Errorf("move %q: unknown category %q", m.Name, lua.LVAsString(entry.RawGetString("category")))
			return
		}

		m.Power = int(lua.LVAsNumber(entry.RawGetString("power")))
		m.Accuracy = int(lua.LVAsNumber(entry.RawGetString("accuracy")))
		m.PP = int(lua.LVAsNumber(entry.RawGetString("pp")))
		m.PPMax = m.PP

		if fn, ok := entry.RawGetString("effect").(*lua.LFunction); ok {
			fnName := "effect_" + lua.LVAsString(key)
			vm.L.SetGlobal(fnName, fn)
			m.EffectFunc = fnName
		}

		vm.Moves[lua.LVAsString(key)] = m
	})

	return outerErr
}

var statusEffectNames = map[string]StatusEffect{
	"burn":       StatusBurn,
	"paralysis":  StatusParalyze,
	"sleep":      StatusSleep,
	"poison":     StatusPoison,
	"freeze":     StatusFreeze,
	"bad_poison": StatusBadPoison,
}

func (vm *VM) loadStatusHooks() error {
	tbl, ok := vm.L.GetGlobal("StatusEffects").(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected StatusEffects to be a table")
	}

	var outerErr error
	tbl.ForEach(func(key, val lua.LValue) {
		if outerErr != nil {
			return
		}
		entry, ok := val.(*lua.LTable)
		if !ok {
			return
		}

		name := lua.LVAsString(key)
		status, ok := statusEffectNames[name]
		if !ok {
			outerErr = fmt.Errorf("unknown status effect %q", name)
			return
		}

		hooks := StatusHooks{}

		if fn, ok := entry.RawGetString("on_turn_start").(*lua.LFunction); ok {
			fnName := "status_" + name + "_on_turn_start"
			vm.L.SetGlobal(fnName, fn)
			hooks.OnTurnStart = fnName
		}
		if fn, ok := entry.RawGetString("on_turn_end").(*lua.LFunction); ok {
			fnName := "status_" + name + "_on_turn_end"
			vm.L.SetGlobal(fnName, fn)
			hooks.OnTurnEnd = fnName
		}

		vm.StatusHooks[status] = hooks
	})

	return outerErr
}

func (vm *VM) MustSpecies(key string) *Species {
	s, ok := vm.Species[key]
	if !ok {
		panic(fmt.Sprintf("unknown species %q", key))
	}
	return s
}

func (vm *VM) MustMove(key string) *Move {
	m, ok := vm.Moves[key]
	if !ok {
		panic(fmt.Sprintf("unknown move %q", key))
	}
	return m
}
