import type { ReactNode } from "react";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

export default async function RootLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ workspace: string }>;
}) {
  const { workspace: slug } = await params;
  const session = await auth();

  const workspaces = await getWorkspaces(session!.token);
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === slug.toLowerCase(),
  );

  if (workspace?.userRole !== "admin") {
    redirect("/settings/account");
  }

  return <>{children}</>;
}
