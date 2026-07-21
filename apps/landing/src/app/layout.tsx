import "../styles/global.css";
import type { Metadata } from "next";
import { type ReactNode } from "react";
import { GoogleAnalytics, GoogleTagManager } from "@next/third-parties/google";
import { cn } from "lib";
import { DEFAULT_SOCIAL_IMAGE, DEFAULT_TWITTER_IMAGE } from "@/lib/seo";
import { mono, sans, serif } from "@/styles/fonts";
import { Toaster } from "./toaster";
import Providers from "./providers";

export const metadata: Metadata = {
  title: "FortyOne | Customer Feedback and Project Management",
  description:
    "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
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
    "customer feedback management",
    "feedback portal",
    "public product roadmap",
    "team objectives",
    "productivity tools",
    "work management",
    "team performance",
    "collaboration platform",
  ],
  openGraph: {
    type: "website",
    locale: "en_US",
    title: "FortyOne | Customer Feedback and Project Management",
    description:
      "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
    images: [DEFAULT_SOCIAL_IMAGE],
    siteName: "FortyOne",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@fortyoneapp",
    creator: "@fortyoneapp",
    title: "FortyOne | Customer Feedback and Project Management",
    description:
      "Collect requests, decide what matters, and move accepted feedback into project plans with clear goals, owners, estimates, schedules, and delivery tracking.",
    images: [DEFAULT_TWITTER_IMAGE],
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
