import type { ReactNode } from "react";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { auth } from "@/auth";

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const headersList = await headers();
  const subdomain = headersList.get("host")!.split(".")[0];
  const session = await auth();

  const workspaces = session?.workspaces || [];
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
  );

  if (!workspace || workspace.userRole === "guest") {
    redirect("/settings/account");
  }

  return <>{children}</>;
}
