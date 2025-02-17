import type { Metadata } from "next";
import { InviteTeam } from "@/modules/onboarding/invite";

export const metadata: Metadata = {
  title: "Invite Team - Complexus",
  description: "Invite your team to Complexus",
};

export default function InvitePage() {
  return <InviteTeam />;
}
