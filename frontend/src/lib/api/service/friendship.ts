import { client } from "../client";
import type { ApiResponse } from "../types";

export const friendshipAPI = {
  /**
   * Sends a friend request to another user.
   * @param addresseeId - The user_id of the user to send the request to
   */
  sendRequest: (addresseeId: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>("/friends/request", {
      method: "POST",
      body: JSON.stringify({ addressee_id: addresseeId }),
    }),

  /**
   * Responds to a pending friend request.
   * @param friendshipId - The ID of the friendship record
   * @param response - The friend request response, either 'accept' or 'decline'
   */
  respondToRequest: (
    friendshipId: string,
    response: "accept" | "decline",
  ): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/friends/request/${friendshipId}`, {
      method: "PUT",
      body: JSON.stringify({ response }),
    }),
};
