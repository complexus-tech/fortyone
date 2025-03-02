import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { auth } from "@/auth";

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  const userRole = session?.user?.userRole;

  if (userRole !== "admin") {
    redirect("/settings");
  }

  return <>{children}</>;
}
