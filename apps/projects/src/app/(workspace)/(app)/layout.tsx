import type { ReactNode } from "react";
import { notFound } from "next/navigation";
import { ApplicationLayout } from "@/components/layouts";
import { auth } from "@/auth";
import { getTeam } from "@/modules/teams/queries";

export default async function RootLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ teamId: string }>;
}) {
  const session = await auth();
  const { teamId } = await params;
  const team = await getTeam(teamId, session!);
  if (!team) {
    notFound();
  }
  return <ApplicationLayout>{children}</ApplicationLayout>;
}
