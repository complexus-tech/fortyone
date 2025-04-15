import type { Session } from "next-auth";
import type { Invitation } from "@/types";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const getRedirectUrl = (
  session: Session,
  invitations: Invitation[] = [],
) => {
  if (session.workspaces.length === 0) {
    if (invitations.length > 0) {
      return `/onboarding/join?token=${invitations[0].token}`;
    }
    return "/onboarding/create";
  }
  const activeWorkspace = session.activeWorkspace || session.workspaces[0];
  if (domain.includes("localhost")) {
    return `http://${activeWorkspace.slug}.localhost:3000/my-work`;
  }
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};

export const buildWorkspaceUrl = (slug: string) => {
  if (domain.includes("localhost")) {
    return `http://${slug}.localhost:3000/my-work`;
  }
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
