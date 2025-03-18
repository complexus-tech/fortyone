import type { Metadata } from "next";
import { InvitationsPage } from "@/modules/settings/invitations";

export const metadata: Metadata = {
  title: "My Invitations",
  description: "Manage your workspace invitations",
};

export default function Page() {
  return <InvitationsPage />;
}
