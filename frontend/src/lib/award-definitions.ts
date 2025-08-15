export interface AwardDefinition {
  title: string;
  description: string;
  calculation?: string;
}

export const AWARD_DEFINITIONS: AwardDefinition[] = [
  {
    title: "The Oracle",
    description: "Awarded for having the most rounds with a correct bid.",
  },
  {
    title: "The Gambler",
    description: "Awarded for having the most rounds with an incorrect bid.",
  },
  {
    title: "The Treasure Hunter",
    description: "Awarded for collecting the highest total bonus points.",
  },
  {
    title: "The Scallywag",
    description: "Awarded for the most successful zero-trick bids.",
  },
  {
    title: "The Buccaneer",
    description: "Awarded for taking the most tricks throughout the game.",
  },
  {
    title: "The Gunslinger",
    description:
      "Awarded for having the highest average bid, regardless of whether the bids were successful.",
  },
  {
    title: "The Conservative",
    description:
      "Awarded for being the most effective at winning with low bids (highest points-per-trick on correct bids).",
    calculation:
      "Calculated by finding the highest 'Points per Trick' ratio, but only for rounds where a player's bid was correct (and greater than zero).",
  },
  {
    title: "The Maverick",
    description:
      "Awarded to the player with the wildest swings in bidding (highest standard deviation).",
    calculation:
      "Calculated by finding the highest standard deviation among all bids made by each player throughout the game. A higher value indicates more inconsistent bidding (e.g., bidding very high in some rounds and very low in others).",
  },
  {
    title: "The Overboard",
    description:
      "Awarded for having the largest single-round difference between a bid and tricks taken.",
  },
  {
    title: "The Wild Card",
    description:
      "Awarded for the most inconsistent performance, based on the highest variance in round scores.",
    calculation:
      "Calculated by finding the highest statistical variance in round scores for each player. A high variance indicates a 'boom-or-bust' performance with very high-scoring and very low-scoring rounds.",
  },
  {
    title: "The Closer",
    description:
      "Awarded for scoring the most points in the final three rounds of the game.",
  },
  {
    title: "The Scoundrel",
    description:
      "Awarded for accumulating the most asterisks for misplays or rule infractions.",
  },
  {
    title: "The Mutineer",
    description: "Awarded for the most failed zero-trick bids.",
  },
  {
    title: "The Anchor",
    description: "Lovingly awarded to the player who finished in last place.",
  },
];
