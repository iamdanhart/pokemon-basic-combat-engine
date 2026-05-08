package engine

import (
	"fmt"
	"math/rand"

	lua "github.com/yuin/gopher-lua"
)

type BattleState int

const (
	StateStart BattleState = iota
	StateChooseMove
	StateResolveTurn
	StateCheckFaint
	StateBattleOver
)

type BattleResult int

const (
	ResultPlayerWin BattleResult = iota
	ResultPlayerLoss
	ResultDraw
)

type Battle struct {
	Player *Pokemon
	Enemy  *Pokemon
	State  BattleState
	Turn   int
	vm     *VM
}

func NewBattle(player, enemy *Pokemon, vm *VM) *Battle {
	return &Battle{
		Player: player,
		Enemy:  enemy,
		State:  StateStart,
		Turn:   1,
		vm:     vm,
	}
}

func (b *Battle) applyMove(actor, target *Pokemon, move *Move) {
	if move.Category == CategoryStatus {
		b.callLuaEffect(actor, target, move)
		return
	}

	// Accuracy == 0 is a sentinel meaning always hits
	if move.Accuracy > 0 && rand.Intn(100)+1 > move.Accuracy {
		fmt.Printf("%s's attack missed!\n", actor.Name)
		return
	}

	base := calcDamage(actor, target, move)

	// Apply type effectiveness across all defender types
	eff := 1.0
	for i := 0; i < target.Species.NumTypes; i++ {
		eff *= Effectiveness(move.Type, target.Species.Types[i])
	}

	damage := int(float64(base) * eff)
	if damage < 1 {
		damage = 1
	}

	if eff > 1.0 {
		fmt.Println("It's super effective!")
	} else if eff < 1.0 && eff > 0 {
		fmt.Println("It's not very effective...")
	} else if eff == 0 {
		fmt.Printf("It doesn't affect %s...\n", target.Name)
		return
	}

	target.HP -= damage
	fmt.Printf("%s took %d damage! %s\n", target.Name, damage, target)
}

// Modified Gen 1 formula - omits the random factor and STAB bonus for simplicity
func calcDamage(attacker, defender *Pokemon, move *Move) int {
	return ((attacker.Level*2/5+2)*move.Power*attacker.Atk/defender.Def)/50 + 2
}

func (b *Battle) callLuaEffect(actor, target *Pokemon, move *Move) {
	if move.EffectFunc == "" {
		return
	}

	actorTbl := pokemonToLuaTable(b.vm.L, actor)
	targetTbl := pokemonToLuaTable(b.vm.L, target)

	err := b.vm.L.CallByParam(lua.P{
		Fn:      b.vm.L.GetGlobal(move.EffectFunc),
		NRet:    0,
		Protect: true,
	}, actorTbl, targetTbl)

	if err != nil {
		fmt.Printf("[lua effect error] %s: %v\n", move.EffectFunc, err)
		return
	}

	syncFromLuaTable(actor, actorTbl)
	syncFromLuaTable(target, targetTbl)
}

func pokemonToLuaTable(L *lua.LState, p *Pokemon) *lua.LTable {
	t := L.NewTable()
	L.SetField(t, "name", lua.LString(p.Name))
	L.SetField(t, "hp", lua.LNumber(p.HP))
	L.SetField(t, "max_hp", lua.LNumber(p.MaxHP))
	L.SetField(t, "atk", lua.LNumber(p.Atk))
	L.SetField(t, "def", lua.LNumber(p.Def))
	L.SetField(t, "spd", lua.LNumber(p.Spd))
	L.SetField(t, "status", lua.LNumber(p.StatusEffect))
	return t
}

func syncFromLuaTable(p *Pokemon, t *lua.LTable) {
	p.HP = int(lua.LVAsNumber(t.RawGetString("hp")))
	p.Atk = int(lua.LVAsNumber(t.RawGetString("atk")))
	p.Def = int(lua.LVAsNumber(t.RawGetString("def")))
	p.Spd = int(lua.LVAsNumber(t.RawGetString("spd")))
	p.StatusEffect = StatusEffect(lua.LVAsNumber(t.RawGetString("status")))
}

func (b *Battle) ResolveTurn(playerMove, enemyMove *Move) {
	first, firstMove, second, secondMove := b.Player, playerMove, b.Enemy, enemyMove
	if b.Enemy.Spd > b.Player.Spd || (b.Enemy.Spd == b.Player.Spd && rand.Intn(2) == 0) {
		first, firstMove, second, secondMove = b.Enemy, enemyMove, b.Player, playerMove
	}

	fmt.Printf("\n%s used %s!\n", first.Name, firstMove.Name)
	b.applyMove(first, second, firstMove)

	if !second.IsFainted() {
		fmt.Printf("%s used %s!\n", second.Name, secondMove.Name)
		b.applyMove(second, first, secondMove)
	}
}

func (b *Battle) checkFaints() BattleResult {
	playerFainted := b.Player.IsFainted()
	enemyFainted := b.Enemy.IsFainted()

	if playerFainted && enemyFainted {
		return ResultDraw
	}
	if playerFainted {
		return ResultPlayerLoss
	}
	if enemyFainted {
		return ResultPlayerWin
	}
	return -1
}

func (b *Battle) enemyChooseMove() *Move {
	for i := range b.Enemy.Moves {
		return &b.Enemy.Moves[i]
	}
	return nil
}
