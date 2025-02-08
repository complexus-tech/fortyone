import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Box } from "ui";
import { SessionProvider } from "next-auth/react";
import { auth } from "@/auth";
import { Navigation } from "./navigation";

export const metadata: Metadata = {
  title: "Onboarding | Complexus",
  description: "Set up your Complexus workspace and invite your team.",
};

export default async function OnboardingLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  return (
    <SessionProvider session={session}>
      <Box className="relative">
        <Navigation />
        {children}
      </Box>
    </SessionProvider>
  );
}
