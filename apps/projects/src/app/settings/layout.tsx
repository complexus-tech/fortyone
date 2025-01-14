import type { Metadata } from "next";
import type { ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { SettingsLayout } from "@/components/layouts";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Settings",
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  return (
    <SessionProvider session={session}>
      <SettingsLayout>{children}</SettingsLayout>
    </SessionProvider>
  );
}
