/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { QueryClient } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { invitationKeys } from "../keys";
import type { Invitation } from "../types";
import {
  optimisticallyAcceptInvitation,
  optimisticallyRevokeInvitation,
  reconcileAcceptedInvitation,
  reconcileRevokedInvitation,
  rollbackAcceptedInvitation,
  rollbackRevokedInvitation,
} from "./cache";

const createInvitation = (
  id: string,
  overrides: Partial<Invitation> = {},
): Invitation => ({
  createdAt: "2026-08-29T08:00:00.000Z",
  email: `${id}@example.com`,
  expiresAt: "2026-08-30T08:00:00.000Z",
  id,
  inviterId: "inviter-1",
  role: "member",
  teamIds: [],
  token: `token-${id}`,
  updatedAt: "2026-08-29T08:00:00.000Z",
  workspaceColor: "#111111",
  workspaceId: "workspace-1",
  workspaceName: "Complexus",
  workspaceSlug: "complexus",
  ...overrides,
});

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

describe("invitation optimistic cache policy", () => {
  it("removes an accepted invitation by token and restores the exact snapshot", async () => {
    const queryClient = createQueryClient();
    const selectedInvitation = createInvitation("invitation-1");
    const remainingInvitation = createInvitation("invitation-2");
    const previousInvitations = [selectedInvitation, remainingInvitation];
    const cancelQueries = jest.spyOn(queryClient, "cancelQueries");
    queryClient.setQueryData(invitationKeys.mine, previousInvitations);

    const context = await optimisticallyAcceptInvitation(
      queryClient,
      selectedInvitation.token!,
    );

    expect(cancelQueries).toHaveBeenCalledWith({
      queryKey: invitationKeys.mine,
    });
    expect(context).toEqual({
      invitation: selectedInvitation,
      previousMineInvitations: previousInvitations,
    });
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual([
      remainingInvitation,
    ]);

    rollbackAcceptedInvitation(queryClient, context);
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual(
      previousInvitations,
    );
  });

  it("reconciles accepted invitations with personal and workspace caches", () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");

    reconcileAcceptedInvitation(queryClient);

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: invitationKeys.mine,
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: workspaceKeys.lists(),
    });
  });

  it("removes a revoked invitation from both views and restores both snapshots", async () => {
    const queryClient = createQueryClient();
    const selectedInvitation = createInvitation("invitation-1");
    const pendingInvitation = createInvitation("invitation-2");
    const personalInvitation = createInvitation("invitation-3");
    const pendingInvitations = [selectedInvitation, pendingInvitation];
    const mineInvitations = [selectedInvitation, personalInvitation];
    queryClient.setQueryData(
      invitationKeys.pending("complexus"),
      pendingInvitations,
    );
    queryClient.setQueryData(invitationKeys.mine, mineInvitations);

    const context = await optimisticallyRevokeInvitation(
      queryClient,
      "complexus",
      selectedInvitation.id,
    );

    expect(
      queryClient.getQueryData(invitationKeys.pending("complexus")),
    ).toEqual([pendingInvitation]);
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual([
      personalInvitation,
    ]);

    rollbackRevokedInvitation(queryClient, "complexus", context);
    expect(
      queryClient.getQueryData(invitationKeys.pending("complexus")),
    ).toEqual(pendingInvitations);
    expect(queryClient.getQueryData(invitationKeys.mine)).toEqual(
      mineInvitations,
    );
  });

  it("reconciles only the revoked invitation workspace and personal cache", () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");

    reconcileRevokedInvitation(queryClient, "complexus");

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: invitationKeys.pending("complexus"),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: invitationKeys.mine,
    });
    expect(invalidateQueries).not.toHaveBeenCalledWith({
      queryKey: invitationKeys.pending("other-workspace"),
    });
  });
});
