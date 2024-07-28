import type { Metadata } from "next";
import type { ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { ApplicationLayout } from "@/components/layouts";
import { StoreProvider } from "@/context/store";
import { getStates } from "@/lib/queries/states/get-states";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Objectives",
  description: "Complexus Objectives",
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const [states, objectives, teams, sprints] = await Promise.all([
    getStates(),
    getObjectives(),
    getTeams(),
    getSprints(),
  ]);

  const session = await auth();

  return (
    <SessionProvider session={session}>
      <StoreProvider initialState={{ states, objectives, teams, sprints }}>
        <ApplicationLayout>{children}</ApplicationLayout>
      </StoreProvider>
    </SessionProvider>
  );
}
