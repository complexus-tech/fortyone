import type { Invitation, Workspace } from "@/types";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const getRedirectUrl = (
  workspaces: Workspace[],
  invitations: Invitation[] = [],
  lastUsedWorkspaceId?: string,
) => {
  if (workspaces.length === 0) {
    if (invitations.length > 0) {
      return `/onboarding/join?token=${invitations[0].token}`;
    }
    return "/onboarding/create";
  }
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ||
    workspaces[0];
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};

export const buildWorkspaceUrl = (slug: string) => {
  return `https://${slug}.${domain}/my-work`;
};

const second = 1000;
const minute = 60 * second;
const hour = 60 * minute;
const day = 24 * hour;
const week = 7 * day;
const month = 30 * day;
const year = 365 * day;

export const DURATION_FROM_MILLISECONDS = {
  SECOND: second,
  MINUTE: minute,
  HOUR: hour,
  DAY: day,
  WEEK: week,
  MONTH: month,
  YEAR: year,
};

export const DURATION_FROM_SECONDS = {
  SECOND: second / second,
  MINUTE: minute / second,
  HOUR: hour / second,
  DAY: day / second,
  WEEK: week / second,
  MONTH: month / second,
  YEAR: year / second,
};
