import { queryOptions } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { AuthHeaderOptions } from "@/lib/http/auth-headers";
import type { WorkspaceCtx } from "@/lib/http";
import { invitationKeys } from "../keys";
import { getMyInvitations } from "./my-invitations";
import { getPendingInvitations } from "./pending-invitations";

const MY_INVITATIONS_STALE_TIME = DURATION_FROM_MILLISECONDS.MINUTE * 10;
const INVITATION_PREFETCH_STALE_TIME = DURATION_FROM_MILLISECONDS.MINUTE * 5;

const createMyInvitationsOptions = (
  authOptions: AuthHeaderOptions,
  staleTime: number,
) =>
  queryOptions({
    queryKey: invitationKeys.mine,
    queryFn: () => getMyInvitations(authOptions),
    staleTime,
  });

const createPendingInvitationsOptions = (
  ctx: WorkspaceCtx,
  staleTime?: number,
) =>
  queryOptions({
    queryKey: invitationKeys.pending(ctx.workspaceSlug),
    queryFn: () => getPendingInvitations(ctx),
    staleTime,
  });

export const myInvitationsQueryOptions = () =>
  createMyInvitationsOptions({}, MY_INVITATIONS_STALE_TIME);

export const pendingInvitationsQueryOptions = (ctx: WorkspaceCtx) =>
  createPendingInvitationsOptions(ctx);

export const myInvitationsPrefetchOptions = (authOptions: AuthHeaderOptions) =>
  createMyInvitationsOptions(authOptions, INVITATION_PREFETCH_STALE_TIME);

export const pendingInvitationsPrefetchOptions = (ctx: WorkspaceCtx) =>
  createPendingInvitationsOptions(ctx, INVITATION_PREFETCH_STALE_TIME);
