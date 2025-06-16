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

  /**
   * Removes a friendship with another user.
   * @param friendId - The user_id of the user to unfriend
   */
  unfriend: (friendId: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/friends/${friendId}`, {
      method: "DELETE",
    }),

  /**
   * Cancels a friend request sent to another user.
   * @param addresseeId - The user_id of the user the request was sent to
   */
  cancelRequest: (addresseeId: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/friends/request/cancel/${addresseeId}`, {
      method: "DELETE",
    }),

  /**
   * Blocks another user.
   * @param userIdToBlock - The user_id of the user to block
   */
  blockUser: (userIdToBlock: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/friends/block/${userIdToBlock}`, {
      method: "POST",
    }),

  /**
   * Unblocks another user.
   * @param userIdToUnblock - The user_id of the user to unblock
   */
  unblockUser: (userIdToUnblock: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/friends/block/${userIdToUnblock}`, {
      method: "DELETE",
    }),
};
