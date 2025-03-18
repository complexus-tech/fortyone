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
  statusKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import { getMyInvitations } from "@/modules/invitations/queries/my-invitations";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getUnreadNotifications } from "@/modules/notifications/queries/get-unread";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getProfile } from "@/lib/queries/users/profile";
import { getMembers } from "@/lib/queries/members/get-members";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

export const fetchNonCriticalImportantQueries = (
  queryClient: QueryClient,
  token: string,
) => {
  queryClient.prefetchQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getAutomationPreferences(),
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.public(),
    queryFn: getPublicTeams,
  });
  queryClient.prefetchQuery({
    queryKey: sprintKeys.lists(),
    queryFn: getSprints,
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: getObjectiveStatuses,
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.list(),
    queryFn: getObjectives,
  });
  queryClient.prefetchQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(),
  });
  queryClient.prefetchQuery({
    queryKey: invitationKeys.mine,
    queryFn: getMyInvitations,
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.unread(),
    queryFn: getUnreadNotifications,
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getWorkspaceSettings(),
  });
  queryClient.prefetchQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(),
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(),
    queryFn: getMembers,
  });
  queryClient.prefetchQuery({
    queryKey: statusKeys.lists(),
    queryFn: getStatuses,
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getWorkspaces(token),
  });
  return queryClient;
};
