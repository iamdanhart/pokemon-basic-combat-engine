# Pokemon Basic Combat Engine

A simplified Pokemon-style combat engine built in Go with Lua scripting. This is a learning project - the goal is to understand how a basic game engine is structured and how a scripting language integrates with a compiled runtime.

## Scope

- Single battles between two Pokemon
- Original 151 species
- Modern type mechanics (all 18 types including Fairy)
- Move effects and status conditions scripted in Lua
- No multi-Pokemon parties, held items, in-battle items, or progression system
- Stats simplified to HP, Atk, Def, and Spd - SpAtk, SpDef, EVs, and IVs omitted
- No critical hits
- No priority moves (e.g. Quick Attack) - speed always determines turn order
- No stat stages (moves like Growl or Sand Attack that modify stat/accuracy multipliers)
- No field conditions (e.g. weather, Trick Room, Gravity) that alter battle mechanics
- No abilities or natures
- No multi-hit moves (e.g. Fury Attack)
- No two-turn moves (e.g. Fly, Dig)
- No recoil moves (e.g. Double-Edge) — Struggle is the exception and does deal recoil
- No binding moves (e.g. Wrap)
- No self-destruct moves (e.g. Explosion)

## Conventions

- Move `accuracy` of `0` is a sentinel meaning the move always hits - no accuracy check is performed
- Damage uses a modified Gen 1 formula - the random factor and STAB bonus are omitted for simplicity. Moves with custom damage calculations (e.g. Psywave) are not supported.
- Type effectiveness is stored as integers scaled by 10 (20=2x, 10=1x, 5=0.5x, 0=immune) to avoid floating-point imprecision in the chart. `Effectiveness()` converts to `float64` at the call site.

## Architecture

The engine is split into two layers:

- **Go** - battle loop, damage calculation, type effectiveness, state machine
- **Lua** - species data, move data, move effects, status effect hooks

Species and move data are defined in `scripts/` as Lua tables and loaded at startup. Non-damaging move effects and status condition behavior (burn, paralysis, etc.) are scripted as Lua callbacks, keeping content changes out of the compiled engine.

## Known Issues

- **Battle is hardcoded.** `main.go` always starts Bulbasaur vs Charmander at level 5 with fixed moves. There is no species or move selection.
- **Type-ahead input is not discarded between turns.** Input typed during turn resolution is buffered and consumed as the next move choice. Invalid input is rejected and the player is re-prompted, so this is recoverable but may be surprising.

## Running

```bash
go run .
```

## Testing

```bash
go test ./...
```