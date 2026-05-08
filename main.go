package main

import (
	"log"

	"github.com/iamdanhart/pokemon-basic-combat-engine/engine"
)

func main() {
	vm := engine.NewVM()
	defer vm.Close()

	if err := vm.LoadScripts("scripts"); err != nil {
		log.Fatal(err)
	}

	player := engine.NewPokemon(vm.MustSpecies("bulbasaur"), 5, []engine.Move{
		*vm.MustMove("vine_whip"),
		*vm.MustMove("thunder_wave"),
	})
	enemy := engine.NewPokemon(vm.MustSpecies("charmander"), 5, []engine.Move{
		*vm.MustMove("ember"),
		*vm.MustMove("tackle"),
	})

	battle := engine.NewBattle(player, enemy, vm)
	battle.Run()
}