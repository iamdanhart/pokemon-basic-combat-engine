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
            if target.status == Status.none then
                target.status = Status.paralysis
            end
        end,
    },
    scratch = {
        name = "Scratch",
        type = "normal",
        category = "physical",
        power = 40, accuracy = 100, pp = 35,
    },
    slash = {
        name = "Slash",
        type = "normal",
        category = "physical",
        power = 70, accuracy = 100, pp = 20,
    },
    flamethrower = {
        name = "Flamethrower",
        type = "fire",
        category = "special",
        power = 90, accuracy = 100, pp = 15,
    },
    surf = {
        name = "Surf",
        type = "water",
        category = "special",
        power = 90, accuracy = 100, pp = 15,
    },
    thunderbolt = {
        name = "Thunderbolt",
        type = "electric",
        category = "special",
        power = 90, accuracy = 100, pp = 15,
    },
    ice_beam = {
        name = "Ice Beam",
        type = "ice",
        category = "special",
        power = 90, accuracy = 100, pp = 10,
    },
    will_o_wisp = {
        name = "Will-O-Wisp",
        type = "fire",
        category = "status",
        power = 0, accuracy = 85, pp = 15,
        effect = function(actor, target)
            if target.status == Status.none then
                target.status = Status.burn
            end
        end,
    },
    toxic = {
        name = "Toxic",
        type = "poison",
        category = "status",
        power = 0, accuracy = 90, pp = 10,
        effect = function(actor, target)
            if target.status == Status.none then
                target.status = Status.bad_poison
                target.status_turns = 1
            end
        end,
    },
    poison_powder = {
        name = "PoisonPowder",
        type = "poison",
        category = "status",
        power = 0, accuracy = 75, pp = 35,
        effect = function(actor, target)
            if target.status == Status.none then
                target.status = Status.poison
            end
        end,
    },
}