package engine

import "fmt"

type StatusEffect int

const (
	StatusNone StatusEffect = iota
	StatusBurn
	StatusParalyze
	StatusSleep
	StatusPoison
	StatusFreeze
)

type Species struct {
	Name     string
	Types    [2]Type
	NumTypes int
	BaseHP   int
	BaseAtk  int
	BaseDef  int
	BaseSpd  int
}

type Pokemon struct {
	Species      *Species // pointer so multiple instances share one Species without copying
	Level        int
	Name         string
	MaxHP        int
	HP           int
	Atk          int
	Def          int
	Spd          int
	Moves        []Move
	StatusEffect StatusEffect
	StatusTurns  int // turns remaining for timed statuses (sleep, freeze)
}

func NewPokemon(s *Species, level int, moves []Move) *Pokemon {
	hp := (s.BaseHP * 2 * level / 100) + level + 10 // matches real formula in RBY
	stat := func(base int) int { return base*2*level/100 + 5 }
	return &Pokemon{
		Species: s,
		Level:   level,
		Name:    s.Name,
		MaxHP:   hp,
		HP:      hp,
		Atk:     stat(s.BaseAtk),
		Def:     stat(s.BaseDef),
		Spd:     stat(s.BaseSpd),
		Moves:   moves,
	}
}

func (s StatusEffect) String() string {
	switch s {
	case StatusBurn:
		return "burned"
	case StatusParalyze:
		return "paralyzed"
	case StatusSleep:
		return "asleep"
	case StatusPoison:
		return "poisoned"
	case StatusFreeze:
		return "frozen"
	default:
		return ""
	}
}

func (p *Pokemon) IsFainted() bool {
	return p.HP <= 0
}

func (p *Pokemon) String() string {
	hp := max(p.HP, 0)
	return fmt.Sprintf("%s [%d/%d HP]", p.Name, hp, p.MaxHP)
}
