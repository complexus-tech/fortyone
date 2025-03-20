import type { QueryClient } from "@tanstack/react-query";
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
import {
  getCachedObjectives,
  getCachedSprints,
  getCachedLabels,
  getCachedPublicTeams,
  getCachedMyInvitations,
  getCachedAutomationPreferences,
  getCachedUnreadNotifications,
  getCachedWorkspaceSettings,
  getCachedProfile,
  getCachedMembers,
  getCachedWorkspaces,
} from "@/lib/cached-queries";

export const fetchNonCriticalImportantQueries = (
  queryClient: QueryClient,
  token: string,
) => {
  queryClient.prefetchQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getCachedAutomationPreferences(),
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.public(),
    queryFn: () => getCachedPublicTeams(),
  });
  queryClient.prefetchQuery({
    queryKey: sprintKeys.lists(),
    queryFn: () => getCachedSprints(),
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.list(),
    queryFn: () => getCachedObjectives(),
  });
  queryClient.prefetchQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getCachedLabels(),
  });
  queryClient.prefetchQuery({
    queryKey: invitationKeys.mine,
    queryFn: () => getCachedMyInvitations(),
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.unread(),
    queryFn: () => getCachedUnreadNotifications(),
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getCachedWorkspaceSettings(),
  });
  queryClient.prefetchQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getCachedProfile(),
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(),
    queryFn: () => getCachedMembers(),
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getCachedWorkspaces(token),
  });
  return queryClient;
};
