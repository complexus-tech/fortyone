import { createApiClient } from "api-client";
import { getApiUrl } from "@/lib/api-url";

export type AccountDeletionResult = "deleted" | "cleanup_pending";

export async function deleteAccount(
  expectedUserId: string,
): Promise<AccountDeletionResult> {
  if (!expectedUserId)
    throw new Error(
      "Your account could not be verified. Refresh and try again.",
    );
  // Account deletion is deliberately not retried: a lost success response
  // must not replay a destructive action against a later browser session.
  const response = await createApiClient(getApiUrl()).delete("users/account", {
    retry: 0,
    json: { expectedUserId },
  });
  if (response.status === 204) return "deleted";
  if (response.status === 202) {
    const body: unknown = await response.json();
    if (
      body &&
      typeof body === "object" &&
      "data" in body &&
      body.data &&
      typeof body.data === "object" &&
      "status" in body.data &&
      body.data.status === "cleanup_pending"
    )
      return "cleanup_pending";
  }
  throw new Error(
    "The server did not confirm account deletion. Please refresh before trying again.",
  );
}
