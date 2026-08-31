import type { CurrentUser } from "./types";

export class AuthContractError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "AuthContractError";
  }
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const isString = (value: unknown): value is string => typeof value === "string";

type CurrentUserPayload = Omit<CurrentUser, "lastUsedWorkspaceId"> & {
  lastUsedWorkspaceId: string | null;
};

const isCurrentUser = (value: unknown): value is CurrentUserPayload => {
  if (!isRecord(value)) return false;

  return (
    isString(value.id) &&
    isString(value.username) &&
    isString(value.email) &&
    isString(value.fullName) &&
    (value.avatarUrl === null || isString(value.avatarUrl)) &&
    typeof value.isActive === "boolean" &&
    typeof value.isInternal === "boolean" &&
    (value.lastUsedWorkspaceId === null ||
      isString(value.lastUsedWorkspaceId)) &&
    typeof value.hasSeenWalkthrough === "boolean" &&
    isString(value.timezone) &&
    isString(value.createdAt) &&
    isString(value.updatedAt)
  );
};

export const decodeCurrentUserResponse = (value: unknown) => {
  if (!isRecord(value) || !("data" in value)) {
    throw new AuthContractError("Current-user response is missing data");
  }

  if (value.data === null) return { data: null };

  if (!isCurrentUser(value.data)) {
    throw new AuthContractError("Current-user response has an invalid user");
  }

  return {
    data: {
      ...value.data,
      // The API legitimately returns null before a user joins or creates a
      // workspace. Keep the existing frontend session contract deterministic.
      lastUsedWorkspaceId: value.data.lastUsedWorkspaceId ?? "",
    },
  };
};
