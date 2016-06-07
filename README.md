Smash The Code - Submission Entry
Rank: 410/2,493

Uses a monte carlo type sampling of a few thousand combinations and apply heuristics to each one. The highest scoring initial node is chosen.

There is a lot of room for optimisations.

Initially I was alternating between perform a full search upto 3 nodes deep and then performing random samplings in order to increase the chances of a good play.
It easily rules out certain combinations which would not be helpful in the slightest.

I wrote the flood fill algorithm very quickly. With more research and time, I could have implemented something a bit quicker.

More contiguous data usage...?
