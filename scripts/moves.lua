-- accuracy: 0 means always hits (no accuracy check performed by the engine)
Moves = {
    tackle = {
        name = "Tackle",
        type = "normal",
        category = "physical",
        power = 40, accuracy = 100, pp = 35,
    },
    vine_whip = {
        name = "Vine Whip",
        type = "grass",
        category = "physical",
        power = 45, accuracy = 100, pp = 25,
    },
    ember = {
        name = "Ember",
        type = "fire",
        category = "special",
        power = 40, accuracy = 100, pp = 25,
    },
    water_gun = {
        name = "Water Gun",
        type = "water",
        category = "special",
        power = 40, accuracy = 100, pp = 25,
    },
    thundershock = {
        name = "Thundershock",
        type = "electric",
        category = "special",
        power = 40, accuracy = 100, pp = 30,
    },
    thunder_wave = {
        name = "Thunder Wave",
        type = "electric",
        category = "status",
        power = 0, accuracy = 90, pp = 20,
        effect = function(actor, target)
            if target.status == 0 then
                target.status = 2  -- StatusParalyze
            end
        end,
    },
}