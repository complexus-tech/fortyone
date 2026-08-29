import "server-only";

import { getCookieHeader } from "@/lib/http/header";
import { getMyInvitations } from "./my-invitations";

export const getMyInvitationsForCurrentRequest = async () => {
  return getMyInvitations({ cookieHeader: await getCookieHeader() });
};
