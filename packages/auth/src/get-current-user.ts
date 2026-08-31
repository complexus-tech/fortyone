import { ApiError, get } from "api-client";
import {
  AuthContractError,
  decodeCurrentUserResponse,
} from "./current-user-contract";
import type { ApiResponse, CurrentUser } from "./types";

export class CurrentUserLookupError extends Error {
  readonly status?: number;

  constructor(message: string, status?: number, options?: ErrorOptions) {
    super(message, options);
    this.name = "CurrentUserLookupError";
    this.status = status;
  }
}

export const fetchCurrentUser = async () => {
  try {
    const response = await get<ApiResponse<CurrentUser>>("auth/me");
    const user = decodeCurrentUserResponse(response).data;

    if (!user) {
      throw new AuthContractError(
        "Authenticated current-user response did not include a user",
      );
    }

    return user;
  } catch (cause) {
    if (cause instanceof AuthContractError) throw cause;

    if (cause instanceof ApiError) {
      // ApiError retains the raw response body in `data`. Preserve the status
      // needed by callers without allowing body values into error monitoring.
      throw new CurrentUserLookupError(
        `Current-user lookup failed with status ${cause.status}`,
        cause.status,
      );
    }

    throw new CurrentUserLookupError("Current-user lookup failed", undefined, {
      cause,
    });
  }
};

export const getCurrentUser = async () => {
  try {
    return await fetchCurrentUser();
  } catch (error) {
    if (error instanceof CurrentUserLookupError && error.status === 401) {
      return null;
    }
    throw error;
  }
};
