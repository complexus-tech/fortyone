import "ui/styles.css";
import "../styles/global.css";
import type { Metadata, Viewport } from "next";
import {
  Bricolage_Grotesque as BricolageGrotesque,
  Geist,
} from "next/font/google";
import { type ReactNode } from "react";
import { redirect } from "next/navigation";
import { cn } from "lib";
import { PublicEnv } from "@/public-env";
import { auth } from "@/auth";
import { AdminShell } from "@/components/admin-shell";
import { getProjectsUrl } from "@/lib/env";
import { Providers } from "./providers";
import { Toaster } from "./toaster";

const font = Geist({
  subsets: ["latin"],
  variable: "--font-body",
  display: "swap",
  weight: "variable",
});

const heading = BricolageGrotesque({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});

export const metadata: Metadata = {
  title: "FortyOne Admin",
  description: "Internal administration for FortyOne.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  minimumScale: 1,
  userScalable: false,
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  const projectsUrl = getProjectsUrl();

  if (!session) {
    redirect(projectsUrl);
  }

  if (!session.user.isInternal) {
    redirect(`${projectsUrl}/unauthorized`);
  }

  return (
    <html
      className={cn(font.variable, heading.variable)}
      lang="en"
      suppressHydrationWarning
    >
      <body>
        <PublicEnv />
        <Providers>
          <AdminShell session={session}>{children}</AdminShell>
          <span className="text-icon" />
          <Toaster />
        </Providers>
      </body>
    </html>
  );
}
