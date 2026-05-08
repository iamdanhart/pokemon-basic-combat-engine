package engine

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"

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
	ResultAborted
)

type Battle struct {
	Player  *Pokemon
	Enemy   *Pokemon
	State   BattleState
	Turn    int
	vm      *VM
	scanner *bufio.Scanner
	out     io.Writer
}

var struggleMove = Move{
	Name:     "Struggle",
	Type:     TypeNormal,
	Category: CategoryPhysical,
	Power:    50,
	Accuracy: 0, // always hits
}

func NewBattle(player, enemy *Pokemon, vm *VM) *Battle {
	return &Battle{
		Player:  player,
		Enemy:   enemy,
		State:   StateStart,
		Turn:    1,
		vm:      vm,
		scanner: bufio.NewScanner(os.Stdin),
		out:     os.Stdout,
	}
}

func (b *Battle) applyMove(actor, target *Pokemon, move *Move) {
	if move.PPMax > 0 {
		move.PP--
	}

	if move.Category == CategoryStatus {
		prevStatus := target.StatusEffect
		b.callLuaEffect(actor, target, move)
		if prevStatus == StatusNone && target.StatusEffect != StatusNone {
			fmt.Fprintf(b.out, "%s was %s!\n", target.Name, target.StatusEffect)
		}
		return
	}

	// Accuracy == 0 is a sentinel meaning always hits
	if move.Accuracy > 0 && rand.Intn(100)+1 > move.Accuracy {
		fmt.Fprintf(b.out, "%s's attack missed!\n", actor.Name)
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
		fmt.Fprintln(b.out, "It's super effective!")
	} else if eff < 1.0 && eff > 0 {
		fmt.Fprintln(b.out, "It's not very effective...")
	} else if eff == 0 {
		fmt.Fprintf(b.out, "It doesn't affect %s...\n", target.Name)
		return
	}

	target.HP -= damage
	fmt.Fprintf(b.out, "%s took %d damage! %s\n", target.Name, damage, target)

	if move.PPMax == 0 {
		recoil := max(actor.MaxHP/4, 1)
		actor.HP -= recoil
		fmt.Fprintf(b.out, "%s is hit by recoil! %s\n", actor.Name, actor)
	}
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
		fmt.Fprintf(b.out, "[lua effect error] %s: %v\n", move.EffectFunc, err)
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
	L.SetField(t, "status_turns", lua.LNumber(p.StatusTurns))
	return t
}

func syncFromLuaTable(p *Pokemon, t *lua.LTable) {
	p.HP = int(lua.LVAsNumber(t.RawGetString("hp")))
	p.Atk = int(lua.LVAsNumber(t.RawGetString("atk")))
	p.Def = int(lua.LVAsNumber(t.RawGetString("def")))
	p.Spd = int(lua.LVAsNumber(t.RawGetString("spd")))
	p.StatusEffect = StatusEffect(lua.LVAsNumber(t.RawGetString("status")))
	p.StatusTurns = int(lua.LVAsNumber(t.RawGetString("status_turns")))
}

// callStatusHook calls the named Lua hook for a Pokemon's status effect.
// Returns true if the Pokemon's turn should be skipped.
func (b *Battle) callStatusHook(hookName string, p *Pokemon) bool {
	if hookName == "" {
		return false
	}

	tbl := pokemonToLuaTable(b.vm.L, p)
	b.vm.L.SetField(tbl, "skip_turn", lua.LFalse)

	err := b.vm.L.CallByParam(lua.P{
		Fn:      b.vm.L.GetGlobal(hookName),
		NRet:    0,
		Protect: true,
	}, tbl)

	if err != nil {
		fmt.Fprintf(b.out, "[lua status hook error] %s: %v\n", hookName, err)
		return false
	}

	syncFromLuaTable(p, tbl)
	return tbl.RawGetString("skip_turn") == lua.LTrue
}

func (b *Battle) applyTurnStart(p *Pokemon) bool {
	if p.StatusEffect == StatusNone {
		return false
	}
	hooks := b.vm.StatusHooks[p.StatusEffect]
	return b.callStatusHook(hooks.OnTurnStart, p)
}

func (b *Battle) applyTurnEnd(p *Pokemon) {
	if p.StatusEffect == StatusNone {
		return
	}
	hooks := b.vm.StatusHooks[p.StatusEffect]
	b.callStatusHook(hooks.OnTurnEnd, p)
}

func (b *Battle) ResolveTurn(playerMove, enemyMove *Move) {
	first, firstMove, second, secondMove := b.Player, playerMove, b.Enemy, enemyMove
	if b.Enemy.Spd > b.Player.Spd || (b.Enemy.Spd == b.Player.Spd && rand.Intn(2) == 0) {
		first, firstMove, second, secondMove = b.Enemy, enemyMove, b.Player, playerMove
	}

	if !b.applyTurnStart(first) {
		fmt.Fprintf(b.out, "\n%s used %s!\n", first.Name, firstMove.Name)
		b.applyMove(first, second, firstMove)
	} else {
		fmt.Fprintf(b.out, "\n%s is %s! It can't move!\n", first.Name, first.StatusEffect)
	}

	if !second.IsFainted() {
		if !b.applyTurnStart(second) {
			fmt.Fprintf(b.out, "%s used %s!\n", second.Name, secondMove.Name)
			b.applyMove(second, first, secondMove)
		} else {
			fmt.Fprintf(b.out, "%s is %s! It can't move!\n", second.Name, second.StatusEffect)
		}
	}

	if !b.Player.IsFainted() {
		b.applyTurnEnd(b.Player)
	}
	if !b.Enemy.IsFainted() {
		b.applyTurnEnd(b.Enemy)
	}
}

func (b *Battle) checkFaints() (BattleResult, bool) {
	playerFainted := b.Player.IsFainted()
	enemyFainted := b.Enemy.IsFainted()

	if playerFainted && enemyFainted {
		return ResultDraw, true
	}
	if playerFainted {
		return ResultPlayerLoss, true
	}
	if enemyFainted {
		return ResultPlayerWin, true
	}
	return 0, false
}

func (b *Battle) playerChooseMove() (*Move, bool) {
	if b.allOutOfPP(b.Player) {
		fmt.Fprintf(b.out, "\n%s has no PP left and must use Struggle!\n", b.Player.Name)
		return &struggleMove, true
	}
	for {
		fmt.Fprintln(b.out, "\nChoose a move:")
		for i, m := range b.Player.Moves {
			fmt.Fprintf(b.out, "  %d. %-16s (%s / %s / Pwr %d / PP %d/%d)\n",
				i+1, m.Name, m.Type, m.Category, m.Power, m.PP, m.PPMax)
		}
		fmt.Fprint(b.out, "> ")

		if !b.scanner.Scan() {
			return nil, false
		}
		input := strings.TrimSpace(b.scanner.Text())
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(b.Player.Moves) {
			fmt.Fprintf(b.out, "Enter a number between 1 and %d.\n", len(b.Player.Moves))
			continue
		}
		move := &b.Player.Moves[n-1]
		if move.PP <= 0 {
			fmt.Fprintln(b.out, "That move has no PP left!")
			continue
		}
		return move, true
	}
}

func (b *Battle) allOutOfPP(p *Pokemon) bool {
	for _, m := range p.Moves {
		if m.PP > 0 {
			return false
		}
	}
	return true
}

func (b *Battle) enemyChooseMove() *Move {
	if b.allOutOfPP(b.Enemy) {
		return &struggleMove
	}
	for range 10 {
		m := &b.Enemy.Moves[rand.Intn(len(b.Enemy.Moves))]
		if m.PP > 0 {
			return m
		}
	}
	for i := range b.Enemy.Moves {
		if b.Enemy.Moves[i].PP > 0 {
			return &b.Enemy.Moves[i]
		}
	}
	return &struggleMove
}

func (b *Battle) Run() BattleResult {
	fmt.Fprintln(b.out, "=== POKEMON BATTLE ===")
	fmt.Fprintf(b.out, "Your %s (Lv.%d) vs Enemy %s (Lv.%d)\n",
		b.Player.Name, b.Player.Level, b.Enemy.Name, b.Enemy.Level)

	for {
		fmt.Fprintf(b.out, "\n--- Turn %d ---\n", b.Turn)
		fmt.Fprintf(b.out, "Your:  %s\n", b.Player)
		fmt.Fprintf(b.out, "Enemy: %s\n", b.Enemy)

		playerMove, ok := b.playerChooseMove()
		if !ok {
			fmt.Fprintln(b.out, "\nBattle aborted.")
			return ResultAborted
		}
		enemyMove := b.enemyChooseMove()

		b.ResolveTurn(playerMove, enemyMove)
		b.Turn++

		if result, fainted := b.checkFaints(); fainted {
			fmt.Fprintln(b.out)
			switch result {
			case ResultPlayerWin:
				fmt.Fprintf(b.out, "Enemy %s fainted! You win!\n", b.Enemy.Name)
			case ResultPlayerLoss:
				fmt.Fprintf(b.out, "%s fainted! You lose!\n", b.Player.Name)
			case ResultDraw:
				fmt.Fprintln(b.out, "Both Pokemon fainted! It's a draw!")
			case ResultAborted:
				// handled above before ResolveTurn
			}
			return result
		}
	}
}