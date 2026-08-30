/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { invitationKeys } from "../keys";
import { getMyInvitations } from "./my-invitations";
import {
  myInvitationsPrefetchOptions,
  myInvitationsQueryOptions,
  pendingInvitationsPrefetchOptions,
  pendingInvitationsQueryOptions,
} from "./options";
import { getPendingInvitations } from "./pending-invitations";

jest.mock("./my-invitations", () => ({
  getMyInvitations: jest.fn(),
}));

jest.mock("./pending-invitations", () => ({
  getPendingInvitations: jest.fn(),
}));

const getMyInvitationsMock = jest.mocked(getMyInvitations);
const getPendingInvitationsMock = jest.mocked(getPendingInvitations);

const executeQuery = async (queryFn: unknown) => {
  if (typeof queryFn !== "function") {
    throw new Error("Expected invitation options to define a query function");
  }

  return (queryFn as () => Promise<unknown>)();
};

describe("invitation query options", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    getMyInvitationsMock.mockResolvedValue([]);
    getPendingInvitationsMock.mockResolvedValue([]);
  });

  it("keeps personal hook and prefetch requests on one key and request owner", async () => {
    const hookOptions = myInvitationsQueryOptions();
    const authOptions = {
      cookieHeader: "session=fortyone",
      token: "token-1",
    };
    const prefetchOptions = myInvitationsPrefetchOptions(authOptions);

    expect(hookOptions.queryKey).toEqual(invitationKeys.mine);
    expect(prefetchOptions.queryKey).toEqual(invitationKeys.mine);
    expect(hookOptions.staleTime).toBe(DURATION_FROM_MILLISECONDS.MINUTE * 10);
    expect(prefetchOptions.staleTime).toBe(
      DURATION_FROM_MILLISECONDS.MINUTE * 5,
    );

    await executeQuery(hookOptions.queryFn);
    await executeQuery(prefetchOptions.queryFn);

    expect(getMyInvitationsMock).toHaveBeenNthCalledWith(1, {});
    expect(getMyInvitationsMock).toHaveBeenNthCalledWith(2, authOptions);
  });

  it("keeps pending hook and prefetch requests scoped to the same workspace", async () => {
    const ctx = {
      cookieHeader: "session=fortyone",
      session: { token: "token-1" },
      workspaceSlug: "complexus",
    };
    const hookOptions = pendingInvitationsQueryOptions(ctx);
    const prefetchOptions = pendingInvitationsPrefetchOptions(ctx);

    expect(hookOptions.queryKey).toEqual(invitationKeys.pending("complexus"));
    expect(prefetchOptions.queryKey).toEqual(
      invitationKeys.pending("complexus"),
    );
    expect(hookOptions.staleTime).toBeUndefined();
    expect(prefetchOptions.staleTime).toBe(
      DURATION_FROM_MILLISECONDS.MINUTE * 5,
    );

    await executeQuery(hookOptions.queryFn);
    await executeQuery(prefetchOptions.queryFn);

    expect(getPendingInvitationsMock).toHaveBeenNthCalledWith(1, ctx);
    expect(getPendingInvitationsMock).toHaveBeenNthCalledWith(2, ctx);
  });
});
