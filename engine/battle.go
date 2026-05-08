package engine

import (
	"bufio"
	"fmt"
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
	move.PP--

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

func (b *Battle) playerChooseMove() *Move {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nChoose a move:")
		for i, m := range b.Player.Moves {
			fmt.Printf("  %d. %-16s (%s / %s / Pwr %d / PP %d/%d)\n",
				i+1, m.Name, m.Type, m.Category, m.Power, m.PP, m.PPMax)
		}
		fmt.Print("> ")

		if !scanner.Scan() {
			continue
		}
		input := strings.TrimSpace(scanner.Text())
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(b.Player.Moves) {
			fmt.Printf("Enter a number between 1 and %d.\n", len(b.Player.Moves))
			continue
		}
		move := &b.Player.Moves[n-1]
		if move.PP <= 0 {
			fmt.Println("That move has no PP left!")
			continue
		}
		return move
	}
}

func (b *Battle) enemyChooseMove() *Move {
	// try up to 10 times to find a move with PP remaining
	for range 10 {
		m := &b.Enemy.Moves[rand.Intn(len(b.Enemy.Moves))]
		if m.PP > 0 {
			return m
		}
	}
	// fallback: return the first move with PP
	for i := range b.Enemy.Moves {
		if b.Enemy.Moves[i].PP > 0 {
			return &b.Enemy.Moves[i]
		}
	}
	return nil
}

func (b *Battle) Run() BattleResult {
	fmt.Println("=== POKEMON BATTLE ===")
	fmt.Printf("Your %s (Lv.%d) vs Enemy %s (Lv.%d)\n",
		b.Player.Name, b.Player.Level, b.Enemy.Name, b.Enemy.Level)

	for {
		fmt.Printf("\n--- Turn %d ---\n", b.Turn)
		fmt.Printf("Your:  %s\n", b.Player)
		fmt.Printf("Enemy: %s\n", b.Enemy)

		playerMove := b.playerChooseMove()
		enemyMove := b.enemyChooseMove()

		b.ResolveTurn(playerMove, enemyMove)
		b.Turn++

		if result := b.checkFaints(); result != -1 {
			fmt.Println()
			switch result {
			case ResultPlayerWin:
				fmt.Printf("Enemy %s fainted! You win!\n", b.Enemy.Name)
			case ResultPlayerLoss:
				fmt.Printf("%s fainted! You lose!\n", b.Player.Name)
			case ResultDraw:
				fmt.Println("Both Pokemon fainted! It's a draw!")
			}
			return result
		}
	}
}
