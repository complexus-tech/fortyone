import type { QueryClient } from "@tanstack/react-query";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import {
  labelKeys,
  teamKeys,
  sprintKeys,
  userKeys,
  invitationKeys,
  notificationKeys,
  workspaceKeys,
  memberKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import { getMyInvitations } from "@/modules/invitations/queries/my-invitations";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getUnreadNotifications } from "@/modules/notifications/queries/get-unread";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getProfile } from "@/lib/queries/users/profile";
import { getMembers } from "@/lib/queries/members/get-members";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const fetchNonCriticalImportantQueries = (
  queryClient: QueryClient,
  token: string,
) => {
  queryClient.prefetchQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getAutomationPreferences(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.public(),
    queryFn: () => getPublicTeams(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: sprintKeys.lists(),
    queryFn: () => getSprints(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.list(),
    queryFn: () => getObjectives(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: invitationKeys.mine,
    queryFn: () => getMyInvitations(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.unread(),
    queryFn: () => getUnreadNotifications(),
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getWorkspaceSettings(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(),
    queryFn: () => getMembers(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getWorkspaces(token),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
  return queryClient;
};
