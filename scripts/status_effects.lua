-- Hooks are called by the engine at specific points in the turn.
-- on_turn_start: called before the Pokemon moves; set skip_turn=true to prevent acting
-- on_turn_end:   called after both moves resolve; used for damage over time
--
-- Timed statuses (sleep, freeze): the move effect that applies them must set
-- pokemon.status_turns to the desired duration. The hook decrements each turn
-- and clears the status (pokemon.status = 0) when it reaches 0.

StatusEffects = {
    burn = {
        on_turn_end = function(pokemon)
            pokemon.hp = pokemon.hp - math.floor(pokemon.max_hp / 8)
        end,
    },
    poison = {
        on_turn_end = function(pokemon)
            pokemon.hp = pokemon.hp - math.floor(pokemon.max_hp / 16)
        end,
    },
    paralysis = {
        on_turn_start = function(pokemon)
            if math.random(100) <= 25 then
                pokemon.skip_turn = true
            end
        end,
    },
    sleep = {
        on_turn_start = function(pokemon)
            if pokemon.status_turns <= 0 then
                pokemon.status = 0  -- wake up
            else
                pokemon.status_turns = pokemon.status_turns - 1
                pokemon.skip_turn = true
            end
        end,
    },
    freeze = {
        on_turn_start = function(pokemon)
            if math.random(100) <= 10 then
                pokemon.status = 0  -- thaw out
            else
                pokemon.skip_turn = true
            end
        end,
    },
}