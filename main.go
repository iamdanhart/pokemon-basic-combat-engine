package main

import (
	"fmt"
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
	vineWhip := vm.Moves["vine_whip"]
	ember := vm.Moves["ember"]

	player := engine.NewPokemon(bulbasaur, 5, []engine.Move{*vineWhip})
	enemy := engine.NewPokemon(charmander, 5, []engine.Move{*ember})

	battle := engine.NewBattle(player, enemy, vm)

	fmt.Println("=== POKEMON BATTLE ===")
	fmt.Printf("Your:  %s\n", player)
	fmt.Printf("Enemy: %s\n\n", enemy)

	battle.ResolveTurn(vineWhip, ember)
}
