import { client } from "../client";
import type {
  ApiResponse,
  CreateGamePayload,
  GameResponse,
  AddPlayerToGamePayload,
  GamePlayerResponse,
} from "../types";

export const gameAPI = {
  /**
   * Creates a new game.
   * @param payload - The data required to create a new game
   * @returns Promise resolving to the created game data
   */
  createGame: (
    payload: CreateGamePayload,
  ): Promise<ApiResponse<GameResponse>> =>
    client<ApiResponse<GameResponse>>("/games", {
      method: "POST",
      body: JSON.stringify(payload),
      headers: {
        "Content-Type": "application/json",
      },
    }),

  /**
   * Adds a player to an existing game.
   * @param gameId - The ID of the game to add the player to
   * @param payload - The data for the player to be added
   * @returns Promise resolving to the added player's data
   */
  addPlayerToGame: (
    gameId: string,
    payload: AddPlayerToGamePayload,
  ): Promise<ApiResponse<GamePlayerResponse>> =>
    client<ApiResponse<GamePlayerResponse>>(`/games/${gameId}/players`, {
      method: "POST",
      body: JSON.stringify(payload),
      headers: {
        "Content-Type": "application/json",
      },
    }),
};
