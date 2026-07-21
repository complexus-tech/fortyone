import "../styles/global.css";
import type { Metadata } from "next";
import { type ReactNode } from "react";
import { GoogleAnalytics, GoogleTagManager } from "@next/third-parties/google";
import { cn } from "lib";
import { mono, sans, serif } from "@/styles/fonts";
import { Toaster } from "./toaster";
import Providers from "./providers";

export const metadata: Metadata = {
  title: "FortyOne | AI Project Manager for Modern Teams",
  description:
    "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
  metadataBase: new URL("https://www.fortyone.app"),
  keywords: [
    "project management platform",
    "project management software",
    "team project management",
    "agile project management platform",
    "project management tool",
    "okr project management",
    "project management",
    "OKR software",
    "team collaboration",
    "sprint planning",
    "task management",
    "agile project management",
    "goal tracking",
    "kanban boards",
    "strategic planning",
    "project roadmap",
    "team objectives",
    "productivity tools",
    "work management",
    "team performance",
    "collaboration platform",
  ],
  openGraph: {
    type: "website",
    locale: "en_US",
    title: "FortyOne | AI Project Manager for Modern Teams",
    description:
      "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
    siteName: "FortyOne",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@fortyoneapp",
    creator: "@fortyoneapp",
    title: "FortyOne | AI Project Manager for Modern Teams",
    description:
      "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
  },
  alternates: {
    canonical: "https://www.fortyone.app",
    languages: {
      "en-US": "https://www.fortyone.app",
      "x-default": "https://www.fortyone.app",
    },
  },
};
const isProduction = process.env.NODE_ENV === "production";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      className={cn(sans.variable, mono.variable, serif.variable)}
      lang="en"
      suppressHydrationWarning
    >
      <body>
        <Providers>
          {children}
          <span className="text-icon" />
        </Providers>
        <Toaster />
      </body>
      {isProduction ? (
        <>
          <GoogleAnalytics
            gaId={process.env.NEXT_PUBLIC_GOOGLE_ANALYTICS_ID!}
          />
          <GoogleTagManager gtmId="G-TYRV8FKD2E" />
        </>
      ) : null}
    </html>
  );
}
