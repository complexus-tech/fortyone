import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getRedirectUrl } from "@/utils";

export default async function Layout({ children }: { children: ReactNode }) {
  const session = await auth();
  if (session) {
    redirect(getRedirectUrl(session));
  }
  return <>{children}</>;
}
