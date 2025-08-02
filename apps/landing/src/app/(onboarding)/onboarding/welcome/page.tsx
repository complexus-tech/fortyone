import type { Metadata } from "next";
import { Welcome } from "@/modules/onboarding/welcome";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";

export const metadata: Metadata = {
  title: "Welcome - Complexus",
  description: "Welcome to Complexus",
};

export default async function WelcomePage() {
  const session = await auth();
  const [workspaces, profile] = await Promise.all([
    getWorkspaces(session?.token || ""),
    getProfile(session!),
  ]);
  return <Welcome profile={profile} workspaces={workspaces} />;
}
