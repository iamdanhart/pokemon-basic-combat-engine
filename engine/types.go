package engine

import "fmt"

type Type int

const (
	TypeNormal Type = iota
	TypeFire
	TypeWater
	TypeGrass
	TypeElectric
	TypeIce
	TypeFighting
	TypePoison
	TypeGround
	TypeFlying
	TypePsychic
	TypeBug
	TypeRock
	TypeGhost
	TypeDragon
	TypeDark
	TypeSteel
	TypeFairy
	TypeCount // not a real type; equals the total number of types, used to size the effectiveness chart
)

// typeData is the single source of truth for all types.
// The typeNames and typesByName maps are derived from it in init().
var typeData = []struct {
	typ     Type
	lower   string
	display string
}{
	{TypeNormal, "normal", "Normal"},
	{TypeFire, "fire", "Fire"},
	{TypeWater, "water", "Water"},
	{TypeGrass, "grass", "Grass"},
	{TypeElectric, "electric", "Electric"},
	{TypeIce, "ice", "Ice"},
	{TypeFighting, "fighting", "Fighting"},
	{TypePoison, "poison", "Poison"},
	{TypeGround, "ground", "Ground"},
	{TypeFlying, "flying", "Flying"},
	{TypePsychic, "psychic", "Psychic"},
	{TypeBug, "bug", "Bug"},
	{TypeRock, "rock", "Rock"},
	{TypeGhost, "ghost", "Ghost"},
	{TypeDragon, "dragon", "Dragon"},
	{TypeDark, "dark", "Dark"},
	{TypeSteel, "steel", "Steel"},
	{TypeFairy, "fairy", "Fairy"},
}

var typeNames = map[Type]string{}
var typesByName = map[string]Type{}

// typeChart[attacking][defending] scaled by 10: 20=2x, 10=1x, 5=0.5x, 0=immune
var typeChart [TypeCount][TypeCount]int

func init() {
	for _, td := range typeData {
		typeNames[td.typ] = td.display
		typesByName[td.lower] = td.typ
	}

	// Fill chart with 1x (10) as the default
	for a := Type(0); a < TypeCount; a++ {
		for d := Type(0); d < TypeCount; d++ {
			typeChart[a][d] = 10
		}
	}

	// Non-1x entries (modern Gen 6+ chart)
	set := func(atk, def Type, val int) { typeChart[atk][def] = val }

	// Normal
	set(TypeNormal, TypeRock, 5); set(TypeNormal, TypeSteel, 5); set(TypeNormal, TypeGhost, 0)
	// Fire
	set(TypeFire, TypeFire, 5); set(TypeFire, TypeWater, 5); set(TypeFire, TypeRock, 5); set(TypeFire, TypeDragon, 5)
	set(TypeFire, TypeGrass, 20); set(TypeFire, TypeIce, 20); set(TypeFire, TypeBug, 20); set(TypeFire, TypeSteel, 20)
	// Water
	set(TypeWater, TypeWater, 5); set(TypeWater, TypeGrass, 5); set(TypeWater, TypeDragon, 5)
	set(TypeWater, TypeFire, 20); set(TypeWater, TypeGround, 20); set(TypeWater, TypeRock, 20)
	// Grass
	set(TypeGrass, TypeFire, 5); set(TypeGrass, TypeGrass, 5); set(TypeGrass, TypePoison, 5)
	set(TypeGrass, TypeFlying, 5); set(TypeGrass, TypeBug, 5); set(TypeGrass, TypeDragon, 5); set(TypeGrass, TypeSteel, 5)
	set(TypeGrass, TypeWater, 20); set(TypeGrass, TypeGround, 20); set(TypeGrass, TypeRock, 20)
	// Electric
	set(TypeElectric, TypeElectric, 5); set(TypeElectric, TypeGrass, 5); set(TypeElectric, TypeDragon, 5)
	set(TypeElectric, TypeGround, 0)
	set(TypeElectric, TypeFlying, 20); set(TypeElectric, TypeWater, 20)
	// Ice
	set(TypeIce, TypeWater, 5); set(TypeIce, TypeIce, 5); set(TypeIce, TypeSteel, 5); set(TypeIce, TypeFire, 5)
	set(TypeIce, TypeGrass, 20); set(TypeIce, TypeGround, 20); set(TypeIce, TypeFlying, 20); set(TypeIce, TypeDragon, 20)
	// Fighting
	set(TypeFighting, TypePoison, 5); set(TypeFighting, TypeBug, 5); set(TypeFighting, TypePsychic, 5)
	set(TypeFighting, TypeFlying, 5); set(TypeFighting, TypeFairy, 5)
	set(TypeFighting, TypeGhost, 0)
	set(TypeFighting, TypeNormal, 20); set(TypeFighting, TypeIce, 20); set(TypeFighting, TypeRock, 20)
	set(TypeFighting, TypeDark, 20); set(TypeFighting, TypeSteel, 20)
	// Poison
	set(TypePoison, TypePoison, 5); set(TypePoison, TypeGround, 5); set(TypePoison, TypeRock, 5); set(TypePoison, TypeGhost, 5)
	set(TypePoison, TypeSteel, 0)
	set(TypePoison, TypeGrass, 20); set(TypePoison, TypeFairy, 20)
	// Ground
	set(TypeGround, TypeGrass, 5); set(TypeGround, TypeBug, 5)
	set(TypeGround, TypeFlying, 0)
	set(TypeGround, TypeFire, 20); set(TypeGround, TypeElectric, 20); set(TypeGround, TypePoison, 20)
	set(TypeGround, TypeRock, 20); set(TypeGround, TypeSteel, 20)
	// Flying
	set(TypeFlying, TypeElectric, 5); set(TypeFlying, TypeRock, 5); set(TypeFlying, TypeSteel, 5)
	set(TypeFlying, TypeGrass, 20); set(TypeFlying, TypeFighting, 20); set(TypeFlying, TypeBug, 20)
	// Psychic
	set(TypePsychic, TypePsychic, 5); set(TypePsychic, TypeSteel, 5)
	set(TypePsychic, TypeDark, 0)
	set(TypePsychic, TypeFighting, 20); set(TypePsychic, TypePoison, 20)
	// Bug
	set(TypeBug, TypeFire, 5); set(TypeBug, TypeFighting, 5); set(TypeBug, TypeFlying, 5)
	set(TypeBug, TypeGhost, 5); set(TypeBug, TypeSteel, 5); set(TypeBug, TypeFairy, 5)
	set(TypeBug, TypeGrass, 20); set(TypeBug, TypePsychic, 20); set(TypeBug, TypeDark, 20)
	// Rock
	set(TypeRock, TypeFighting, 5); set(TypeRock, TypeGround, 5); set(TypeRock, TypeSteel, 5)
	set(TypeRock, TypeNormal, 20); set(TypeRock, TypeFire, 20); set(TypeRock, TypeIce, 20)
	set(TypeRock, TypeFlying, 20); set(TypeRock, TypeBug, 20)
	// Ghost
	set(TypeGhost, TypeNormal, 0)
	set(TypeGhost, TypeDark, 5)
	set(TypeGhost, TypeGhost, 20); set(TypeGhost, TypePsychic, 20)
	// Dragon
	set(TypeDragon, TypeSteel, 5)
	set(TypeDragon, TypeFairy, 0)
	set(TypeDragon, TypeDragon, 20)
	// Dark
	set(TypeDark, TypeFighting, 5); set(TypeDark, TypeDark, 5); set(TypeDark, TypeFairy, 5)
	set(TypeDark, TypeGhost, 20); set(TypeDark, TypePsychic, 20)
	// Steel
	set(TypeSteel, TypeFire, 5); set(TypeSteel, TypeWater, 5); set(TypeSteel, TypeElectric, 5); set(TypeSteel, TypeSteel, 5)
	set(TypeSteel, TypeRock, 20); set(TypeSteel, TypeIce, 20); set(TypeSteel, TypeFairy, 20)
	// Fairy
	set(TypeFairy, TypeFire, 5); set(TypeFairy, TypePoison, 5); set(TypeFairy, TypeSteel, 5)
	set(TypeFairy, TypeFighting, 20); set(TypeFairy, TypeDragon, 20); set(TypeFairy, TypeDark, 20)
}

func Effectiveness(atk, def Type) float64 {
	return float64(typeChart[atk][def]) / 10.0
}

func TypeFromString(s string) (Type, error) {
	t, ok := typesByName[s]
	if !ok {
		return TypeNormal, fmt.Errorf("unknown type: %q", s)
	}
	return t, nil
}

func (t Type) String() string {
	name, ok := typeNames[t]
	if !ok {
		return "Unknown"
	}
	return name
}
