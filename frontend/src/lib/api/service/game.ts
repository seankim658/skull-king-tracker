import { client } from "../client";
import type {
  ApiResponse,
  CreateGamePayload,
  GameResponse,
  AddPlayerToGamePayload,
  GamePlayerResponse,
  UpdateGameSettingsPayload,
  ScorecardResponse,
  SubmitBidsPayload,
  SubmitTricksPayload,
  ActiveGameResponse,
  PaginatedGameHistoryResponse,
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
    client<GameResponse>("/games", {
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
    client<GamePlayerResponse>(`/games/${gameId}/players`, {
      method: "POST",
      body: JSON.stringify(payload),
      headers: {
        "Content-Type": "application/json",
      },
    }),

  /**
   * Removes a player from a game.
   * @param gameId - The ID of the game
   * @param gamePlayerId - The ID of the
   */
  removePlayerFromGame: (
    gameId: string,
    gamePlayerId: string,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/games/${gameId}/players/${gamePlayerId}`, {
      method: "DELETE",
    }),

  /**
   * Updates a game status to 'active' to begin paly
   * @param gameId - The ID of the game to start
   */
  startGame: (gameId: string): Promise<ApiResponse<null>> =>
    client<null>(`/games/${gameId}/start`, {
      method: "PUT",
    }),

  /**
   * Fetches all players for a specified game.
   * @param gameId - The ID of the game
   */
  getGamePlayers: (
    gameId: string,
  ): Promise<ApiResponse<GamePlayerResponse[]>> =>
    client<GamePlayerResponse[]>(`/games/${gameId}/players`, {
      method: "GET",
    }),

  /**
   * Fetches the details for a specific game.
   * @param gameId - The ID of the game to fetch
   */
  getGameDetails: (gameId: string): Promise<ApiResponse<GameResponse>> =>
    client<GameResponse>(`/games/${gameId}/details`, {
      method: "GET",
    }),

  /**
   * Updates the settings for a game.
   * @param gameId - The ID of the game to update
   * @param payload - The settings data to update
   */
  updateGameSettings: (
    gameId: string,
    payload: UpdateGameSettingsPayload,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/games/${gameId}/settings`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),

  /**
   * Fetches the entire state of a game's scorecard.
   * @param gameId - The ID of the game to fetch the scorecard for
   */
  getScorecardState: (
    gameId: string,
  ): Promise<ApiResponse<ScorecardResponse>> =>
    client<ScorecardResponse>(`/games/${gameId}/scorecard`, { method: "GET" }),

  /**
   * Submits the bids for all player for a specific round.
   * @param gameId - The ID of the game
   * @param roundNumber - The round number for which bids are being submitted
   * @param payload - The bid data for all players
   */
  submitBids: (
    gameId: string,
    roundNumber: number,
    payload: SubmitBidsPayload,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/games/${gameId}/rounds/${roundNumber}/bids`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  /**
   * Submits the tricks taken by all players for a specific round.
   * @param gameId - The ID of the game
   * @param roundNumber - The round number for which tricks are being submitted
   * @param payload - The trick data for all players
   */
  submitTricks: (
    gameId: string,
    roundNumber: number,
    payload: SubmitTricksPayload,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/games/${gameId}/rounds/${roundNumber}/tricks`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  /**
   * Fetches all active games for the current user.
   */
  getActiveGames: (): Promise<ApiResponse<ActiveGameResponse[]>> =>
    client<ActiveGameResponse[]>("/games/active", { method: "GET" }),

  /**
   * Fetches a paginated history of a user's completed games.
   */
  getGameHistory: (
    page: number,
    pageSize: number,
    sorting: { id: string; desc: boolean }[],
    sessionId?: string | null,
  ): Promise<ApiResponse<PaginatedGameHistoryResponse>> => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    if (sorting.length > 0) {
      params.append("sort_by", sorting[0].id);
      params.append("sort_order", sorting[0].desc ? "desc" : "asc");
    }
    if (sessionId) {
      params.append("session_id", sessionId);
    }
    return client<PaginatedGameHistoryResponse>(
      `/games/history?${params.toString()}`,
    );
  },
};
