import type { Metadata } from "next";
import { JoinWorkspace } from "@/modules/onboarding/join";

export const metadata: Metadata = {
  title: "Join Workspace - Complexus",
  description: "Join a workspace",
};

export default function JoinWorkspacePage() {
  return <JoinWorkspace />;
}
