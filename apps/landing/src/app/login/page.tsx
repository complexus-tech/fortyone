import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { LoginPage } from "@/components/pages/login";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const metadata: Metadata = {
  title: "Login",
};

export default async function Page() {
  const session = await auth();
  if (session) {
    // Determine where to redirect based on workspace status
    if (session.workspaces.length === 0) {
      redirect("/onboarding/create");
    }
    const activeWorkspace = session.activeWorkspace || session.workspaces[0];
    redirect(`https://${activeWorkspace.slug}.${domain}`);
  }

  return <LoginPage />;
}
