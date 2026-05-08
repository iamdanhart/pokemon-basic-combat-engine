package engine

import "testing"

func testSpecies() *Species {
	return &Species{
		Name:     "Bulbasaur",
		Types:    [2]Type{TypeGrass, TypePoison},
		NumTypes: 2,
		BaseHP:   45,
		BaseAtk:  49,
		BaseDef:  49,
		BaseSpd:  45,
	}
}

func TestNewPokemon_stats(t *testing.T) {
	p := NewPokemon(testSpecies(), 5, nil)

	// HP: (45*2*5/100) + 5 + 10 = 4 + 5 + 10 = 19
	if p.MaxHP != 19 {
		t.Errorf("MaxHP: got %d, want 19", p.MaxHP)
	}
	if p.HP != p.MaxHP {
		t.Errorf("HP should equal MaxHP at creation, got %d", p.HP)
	}

	// stat: 49*2*5/100 + 5 = 4 + 5 = 9
	if p.Atk != 9 {
		t.Errorf("Atk: got %d, want 9", p.Atk)
	}
	if p.Def != 9 {
		t.Errorf("Def: got %d, want 9", p.Def)
	}

	// spd: 45*2*5/100 + 5 = 4 + 5 = 9
	if p.Spd != 9 {
		t.Errorf("Spd: got %d, want 9", p.Spd)
	}
}

func TestNewPokemon_name(t *testing.T) {
	p := NewPokemon(testSpecies(), 5, nil)
	if p.Name != "Bulbasaur" {
		t.Errorf("got %q, want %q", p.Name, "Bulbasaur")
	}
}

func TestIsFainted(t *testing.T) {
	p := NewPokemon(testSpecies(), 5, nil)

	if p.IsFainted() {
		t.Error("full HP Pokemon should not be fainted")
	}

	p.HP = 0
	if !p.IsFainted() {
		t.Error("Pokemon at 0 HP should be fainted")
	}

	p.HP = -5
	if !p.IsFainted() {
		t.Error("Pokemon at negative HP should be fainted")
	}
}

func TestPokemonString_clampsNegativeHP(t *testing.T) {
	p := NewPokemon(testSpecies(), 5, nil)
	p.HP = -10
	got := p.String()
	want := "Bulbasaur [0/19 HP]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPokemonString_normalHP(t *testing.T) {
	p := NewPokemon(testSpecies(), 5, nil)
	got := p.String()
	want := "Bulbasaur [19/19 HP]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
