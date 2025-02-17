import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { auth } from "@/auth";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export default async function Layout({ children }: { children: ReactNode }) {
  const session = await auth();
  if (session) {
    // Determine where to redirect based on workspace status
    if (session.workspaces.length === 0) {
      redirect("/onboarding/create");
    }
    const activeWorkspace = session.activeWorkspace || session.workspaces[0];
    if (domain.includes("localhost")) {
      redirect(`http://${activeWorkspace.slug}.localhost:3000/my-work`);
    }
    redirect(`https://${activeWorkspace.slug}.${domain}/my-work`);
  }
  return <>{children}</>;
}
