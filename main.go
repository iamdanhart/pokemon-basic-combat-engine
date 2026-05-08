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

	bulbasaur := vm.Species["bulbasaur"]
	charmander := vm.Species["charmander"]

	player := engine.NewPokemon(bulbasaur, 5, []engine.Move{
		*vm.Moves["vine_whip"],
		*vm.Moves["thunder_wave"],
	})
	enemy := engine.NewPokemon(charmander, 5, []engine.Move{
		*vm.Moves["ember"],
		*vm.Moves["tackle"],
	})

	battle := engine.NewBattle(player, enemy, vm)
	battle.Run()
}