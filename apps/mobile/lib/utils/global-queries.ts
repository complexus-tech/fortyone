import { QueryClient } from "@tanstack/react-query";
import { getWorkspaces } from "../queries/get-workspaces";
import {
  workspaceKeys,
  memberKeys,
  objectiveKeys,
  statusKeys,
  subscriptionKeys,
  teamKeys,
  userKeys,
} from "@/constants/keys";
import { getProfile } from "@/modules/users/queries/get-profile";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getSubscription } from "@/lib/queries/get-subscription";
import { getObjectiveStatuses } from "@/modules/objectives/queries/get-objectives";
import { getStatuses } from "@/modules/statuses/queries/get-statuses";
import { getMembers } from "@/modules/members/queries/get-members";

export const fetchGlobalQueries = async (queryClient: QueryClient) => {
  queryClient.prefetchQuery({
    queryKey: userKeys.profile(),
    queryFn: getProfile,
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: getWorkspaces,
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.lists(),
    queryFn: getTeams,
  });
  queryClient.prefetchQuery({
    queryKey: subscriptionKeys.details,
    queryFn: getSubscription,
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(),
    queryFn: getMembers,
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: getObjectiveStatuses,
  });
  queryClient.prefetchQuery({
    queryKey: statusKeys.lists(),
    queryFn: getStatuses,
  });
};
