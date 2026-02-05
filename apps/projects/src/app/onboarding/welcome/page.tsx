import type { Metadata } from "next";
import { Welcome } from "@/modules/onboarding/welcome";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { getCookieHeader } from "@/lib/http/header";

export const metadata: Metadata = {
  title: "Welcome - FortyOne",
  description: "Welcome to FortyOne",
};

export default async function WelcomePage() {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  const [workspaces, profile] = await Promise.all([
    getWorkspaces(session?.token || "", cookieHeader),
    getProfile({ token: session?.token, cookieHeader }),
  ]);
  return <Welcome profile={profile} workspaces={workspaces} />;
}
