import type { Metadata } from "next";
import { Welcome } from "@/modules/onboarding/welcome";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";

export const metadata: Metadata = {
  title: "Welcome - FortyOne",
  description: "Welcome to FortyOne",
};

export default async function WelcomePage({
  searchParams,
}: {
  searchParams: Promise<{ callbackUrl?: string }>;
}) {
  const [{ callbackUrl }, workspaces, profile] = await Promise.all([
    searchParams,
    getWorkspaces(),
    getProfile(),
  ]);
  return (
    <Welcome
      callbackUrl={callbackUrl}
      profile={profile}
      workspaces={workspaces}
    />
  );
}
