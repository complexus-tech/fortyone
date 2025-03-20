"use server";

import { cache } from "react";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import { getMyInvitations } from "@/modules/invitations/queries/my-invitations";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getUnreadNotifications } from "@/modules/notifications/queries/get-unread";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getProfile } from "@/lib/queries/users/profile";
import { getMembers } from "@/lib/queries/members/get-members";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

// Cache critical queries
export const getCachedTeams = cache(getTeams);
export const getCachedStatuses = cache(getStatuses);
export const getCachedObjectiveStatuses = cache(getObjectiveStatuses);

// Cache non-critical important queries
export const getCachedObjectives = cache(getObjectives);
export const getCachedSprints = cache(getSprints);
export const getCachedLabels = cache(getLabels);
export const getCachedPublicTeams = cache(getPublicTeams);
export const getCachedMyInvitations = cache(getMyInvitations);
export const getCachedAutomationPreferences = cache(getAutomationPreferences);
export const getCachedUnreadNotifications = cache(getUnreadNotifications);
export const getCachedWorkspaceSettings = cache(getWorkspaceSettings);
export const getCachedProfile = cache(getProfile);
export const getCachedMembers = cache(getMembers);
export const getCachedWorkspaces = cache((token: string) =>
  getWorkspaces(token),
);
