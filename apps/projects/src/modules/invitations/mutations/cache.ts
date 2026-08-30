import type { QueryClient } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { invitationKeys } from "../keys";
import type { Invitation } from "../types";

export type AcceptInvitationCacheContext = {
  invitation?: Invitation;
  previousMineInvitations?: Invitation[];
};

export type RevokeInvitationCacheContext = {
  previousInvitations?: Invitation[];
  previousMineInvitations?: Invitation[];
};

export const optimisticallyAcceptInvitation = async (
  queryClient: QueryClient,
  inviteToken: string,
): Promise<AcceptInvitationCacheContext> => {
  await queryClient.cancelQueries({ queryKey: invitationKeys.mine });

  const previousMineInvitations = queryClient.getQueryData<Invitation[]>(
    invitationKeys.mine,
  );
  const invitation = previousMineInvitations?.find(
    (candidate) => candidate.token === inviteToken,
  );

  queryClient.setQueryData<Invitation[]>(
    invitationKeys.mine,
    (currentInvitations = []) =>
      currentInvitations.filter((candidate) => candidate.token !== inviteToken),
  );

  return { invitation, previousMineInvitations };
};

export const rollbackAcceptedInvitation = (
  queryClient: QueryClient,
  context?: AcceptInvitationCacheContext,
) => {
  queryClient.setQueryData(
    invitationKeys.mine,
    context?.previousMineInvitations,
  );
};

export const reconcileAcceptedInvitation = (queryClient: QueryClient) => {
  void queryClient.invalidateQueries({ queryKey: invitationKeys.mine });
  void queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
};

export const optimisticallyRevokeInvitation = async (
  queryClient: QueryClient,
  workspaceSlug: string,
  invitationId: string,
): Promise<RevokeInvitationCacheContext> => {
  const pendingInvitationsKey = invitationKeys.pending(workspaceSlug);

  await queryClient.cancelQueries({ queryKey: pendingInvitationsKey });
  await queryClient.cancelQueries({ queryKey: invitationKeys.mine });

  const previousInvitations = queryClient.getQueryData<Invitation[]>(
    pendingInvitationsKey,
  );
  const previousMineInvitations = queryClient.getQueryData<Invitation[]>(
    invitationKeys.mine,
  );

  const removeInvitation = (currentInvitations: Invitation[] = []) =>
    currentInvitations.filter((invitation) => invitation.id !== invitationId);

  queryClient.setQueryData<Invitation[]>(
    pendingInvitationsKey,
    removeInvitation,
  );
  queryClient.setQueryData<Invitation[]>(invitationKeys.mine, removeInvitation);

  return { previousInvitations, previousMineInvitations };
};

export const rollbackRevokedInvitation = (
  queryClient: QueryClient,
  workspaceSlug: string,
  context?: RevokeInvitationCacheContext,
) => {
  queryClient.setQueryData(
    invitationKeys.pending(workspaceSlug),
    context?.previousInvitations,
  );
  queryClient.setQueryData(
    invitationKeys.mine,
    context?.previousMineInvitations,
  );
};

export const reconcileRevokedInvitation = (
  queryClient: QueryClient,
  workspaceSlug: string,
) => {
  void queryClient.invalidateQueries({
    queryKey: invitationKeys.pending(workspaceSlug),
  });
  void queryClient.invalidateQueries({ queryKey: invitationKeys.mine });
};
