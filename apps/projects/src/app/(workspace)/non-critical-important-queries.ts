import type { QueryClient } from "@tanstack/react-query";
import type { Session } from "next-auth";
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
  subscriptionKeys,
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
import { getPendingInvitations } from "@/modules/invitations/queries/pending-invitations";
import { getNotificationPreferences } from "@/modules/notifications/queries/get-preferences";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const fetchNonCriticalImportantQueries = (
  queryClient: QueryClient,
  session: Session,
) => {
  queryClient.prefetchQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getAutomationPreferences(session),
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.public(),
    queryFn: () => getPublicTeams(session),
  });
  queryClient.prefetchQuery({
    queryKey: sprintKeys.lists(),
    queryFn: () => getSprints(session),
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.list(),
    queryFn: () => getObjectives(session),
  });
  queryClient.prefetchQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(session),
  });
  queryClient.prefetchQuery({
    queryKey: invitationKeys.mine,
    queryFn: () => getMyInvitations(session),
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.unread(),
    queryFn: () => getUnreadNotifications(session),
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getWorkspaceSettings(session),
  });
  queryClient.prefetchQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(session),
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(),
    queryFn: () => getMembers(session),
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getWorkspaces(session.token),
  });
  queryClient.prefetchQuery({
    queryKey: invitationKeys.pending,
    queryFn: () => getPendingInvitations(session),
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.preferences(),
    queryFn: () => getNotificationPreferences(session),
  });
  queryClient.prefetchQuery({
    queryKey: subscriptionKeys.details,
    queryFn: () => getSubscription(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
  return queryClient;
};
