package engine

import (
	"io"
	"testing"
)

func testMove(power, accuracy int) Move {
	return Move{
		Name:     "Test Move",
		Type:     TypeNormal,
		Category: CategoryPhysical,
		Power:    power,
		Accuracy: accuracy,
		PP:       10,
		PPMax:    10,
	}
}

func newTestBattle(player, enemy *Pokemon) *Battle {
	return &Battle{
		Player: player,
		Enemy:  enemy,
		State:  StateStart,
		Turn:   1,
		vm:     NewVM(),
		out:    io.Discard,
	}
}

func TestCalcDamage(t *testing.T) {
	s := testSpecies()
	attacker := NewPokemon(s, 50, nil)
	defender := NewPokemon(s, 50, nil)
	move := testMove(40, 0)

	// With equal atk/def the formula is symmetric and deterministic
	got := calcDamage(attacker, defender, &move)
	want := ((attacker.Level*2/5+2)*move.Power*attacker.Atk/defender.Def)/50 + 2
	if got != want {
		t.Errorf("calcDamage = %d, want %d", got, want)
	}
}

func TestApplyMoveReducesHP(t *testing.T) {
	s := testSpecies()
	attacker := NewPokemon(s, 50, nil)
	defender := NewPokemon(s, 50, nil)
	move := testMove(40, 0) // accuracy 0 = always hits

	b := newTestBattle(attacker, defender)
	before := defender.HP
	b.applyMove(attacker, defender, &move)

	if defender.HP >= before {
		t.Errorf("expected HP to decrease from %d, got %d", before, defender.HP)
	}
}

func TestApplyMoveDecrementsPP(t *testing.T) {
	s := testSpecies()
	attacker := NewPokemon(s, 50, nil)
	defender := NewPokemon(s, 50, nil)
	move := testMove(40, 0)

	b := newTestBattle(attacker, defender)
	b.applyMove(attacker, defender, &move)

	if move.PP != 9 {
		t.Errorf("expected PP 9, got %d", move.PP)
	}
}

func TestApplyMoveStruggleNoPPDecrement(t *testing.T) {
	s := testSpecies()
	attacker := NewPokemon(s, 50, nil)
	defender := NewPokemon(s, 50, nil)

	b := newTestBattle(attacker, defender)
	local := struggleMove // copy so package-level var is unaffected
	b.applyMove(attacker, defender, &local)

	if local.PP != 0 {
		t.Errorf("expected Struggle PP to remain 0, got %d", local.PP)
	}
}

func TestApplyMoveSuperEffective(t *testing.T) {
	fire := &Species{Name: "Fire", Types: [2]Type{TypeFire}, NumTypes: 1, BaseHP: 45, BaseAtk: 49, BaseDef: 49, BaseSpd: 45}
	grass := &Species{Name: "Grass", Types: [2]Type{TypeGrass}, NumTypes: 1, BaseHP: 45, BaseAtk: 49, BaseDef: 49, BaseSpd: 45}

	attacker := NewPokemon(fire, 50, nil)
	neutralTarget := NewPokemon(fire, 50, nil)
	superTarget := NewPokemon(grass, 50, nil)

	fireMove := Move{Name: "Ember", Type: TypeFire, Category: CategorySpecial, Power: 40, Accuracy: 0, PP: 10, PPMax: 10}

	b := newTestBattle(attacker, neutralTarget)
	b.applyMove(attacker, neutralTarget, &fireMove)
	neutralDamage := neutralTarget.MaxHP - neutralTarget.HP

	fireMove.PP = 10 // reset PP
	b2 := newTestBattle(attacker, superTarget)
	b2.applyMove(attacker, superTarget, &fireMove)
	superDamage := superTarget.MaxHP - superTarget.HP

	if superDamage <= neutralDamage {
		t.Errorf("super effective damage (%d) should exceed neutral (%d)", superDamage, neutralDamage)
	}
}

func TestCheckFaintsNone(t *testing.T) {
	s := testSpecies()
	b := newTestBattle(NewPokemon(s, 50, nil), NewPokemon(s, 50, nil))
	if _, fainted := b.checkFaints(); fainted {
		t.Error("expected no faint when both Pokemon have full HP")
	}
}

func TestCheckFaintsPlayerWin(t *testing.T) {
	s := testSpecies()
	player := NewPokemon(s, 50, nil)
	enemy := NewPokemon(s, 50, nil)
	enemy.HP = 0

	b := newTestBattle(player, enemy)
	result, fainted := b.checkFaints()
	if !fainted || result != ResultPlayerWin {
		t.Errorf("expected ResultPlayerWin, got fainted=%v result=%v", fainted, result)
	}
}

func TestCheckFaintsPlayerLoss(t *testing.T) {
	s := testSpecies()
	player := NewPokemon(s, 50, nil)
	enemy := NewPokemon(s, 50, nil)
	player.HP = 0

	b := newTestBattle(player, enemy)
	result, fainted := b.checkFaints()
	if !fainted || result != ResultPlayerLoss {
		t.Errorf("expected ResultPlayerLoss, got fainted=%v result=%v", fainted, result)
	}
}

func TestCheckFaintsDraw(t *testing.T) {
	s := testSpecies()
	player := NewPokemon(s, 50, nil)
	enemy := NewPokemon(s, 50, nil)
	player.HP = 0
	enemy.HP = 0

	b := newTestBattle(player, enemy)
	result, fainted := b.checkFaints()
	if !fainted || result != ResultDraw {
		t.Errorf("expected ResultDraw, got fainted=%v result=%v", fainted, result)
	}
}

func TestResolveTurnFasterGoesFirst(t *testing.T) {
	s := testSpecies()
	// Player is much faster; enemy has 1 HP — player should KO it before it moves
	player := NewPokemon(s, 50, nil)
	enemy := NewPokemon(s, 50, nil)
	player.Spd = 999
	enemy.HP = 1

	move := testMove(100, 0)
	player.Moves = []Move{move}
	enemy.Moves = []Move{testMove(100, 0)}

	b := newTestBattle(player, enemy)
	enemyPPBefore := enemy.Moves[0].PP

	b.resolveTurn(&player.Moves[0], &enemy.Moves[0])

	if enemy.Moves[0].PP != enemyPPBefore {
		t.Error("enemy moved despite being KO'd by the faster player")
	}
	if !enemy.IsFainted() {
		t.Error("expected enemy to be fainted")
	}
}

func TestStruggleRecoil(t *testing.T) {
	s := testSpecies()
	attacker := NewPokemon(s, 50, nil)
	defender := NewPokemon(s, 50, nil)

	b := newTestBattle(attacker, defender)
	b.applyMove(attacker, defender, &struggleMove)

	wantRecoil := max(attacker.MaxHP/4, 1)
	if attacker.HP != attacker.MaxHP-wantRecoil {
		t.Errorf("expected attacker HP %d after recoil, got %d", attacker.MaxHP-wantRecoil, attacker.HP)
	}
}

func TestAllOutOfPP(t *testing.T) {
	s := testSpecies()
	p := NewPokemon(s, 50, []Move{
		{PP: 0, PPMax: 10},
		{PP: 0, PPMax: 10},
	})
	b := newTestBattle(p, p)
	if !b.allOutOfPP(p) {
		t.Error("expected allOutOfPP to return true")
	}

	p.Moves[0].PP = 1
	if b.allOutOfPP(p) {
		t.Error("expected allOutOfPP to return false when one move has PP")
	}
}