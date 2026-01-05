import "ui/styles.css";
import "../styles/global.css";
import type { Metadata } from "next";
import { type ReactNode } from "react";
import { GoogleAnalytics, GoogleTagManager } from "@next/third-parties/google";
import { cn } from "lib";
import Script from "next/script";
import { body, heading, mono } from "@/styles/fonts";
import { JsonLd } from "@/components/shared";
import { Toaster } from "./toaster";
import Providers from "./providers";

export const metadata: Metadata = {
  title:
    "FortyOne - The Open Source Agentic Project Management Platform",
  description:
    "FortyOne is the leading open source project management platform with AI-powered OKRs, sprint planning, and team collaboration. Better than Jira, Notion, and Monday. Try free today.",
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
    title:
      "FortyOne - The Open Source Agentic Project Management Platform",
    description:
      "FortyOne is the leading open source project management platform with AI-powered OKRs, sprint planning, and team collaboration. Open source alternative to Jira, Notion, and Monday. Try free today.",
    siteName: "FortyOne",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@fortyoneapp",
    creator: "@fortyoneapp",
    title:
      "FortyOne - The Open Source Agentic Project Management Platform",
    description:
      "FortyOne is the leading open source project management platform with AI-powered OKRs, sprint planning, and team collaboration. Open source alternative to Jira, Notion, and Monday. Try free today.",
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
    <html lang="en" suppressHydrationWarning>
      <Script id="brevo-conversations" strategy="afterInteractive">
        {`
        (function(d, w, c) {
            w.BrevoConversationsID = '6834856b58b6d2f7800e0e5e';
            w[c] = w[c] || function() {
                (w[c].q = w[c].q || []).push(arguments);
            };
            var s = d.createElement('script');
            s.async = true;
            s.src = 'https://conversations-widget.brevo.com/brevo-conversations.js';
            if (d.head) d.head.appendChild(s);
        })(document, window, 'BrevoConversations');
      `}
      </Script>
      <head>
        <JsonLd />
      </head>
      <body
        className={cn(
          body.variable,
          heading.variable,
          mono.variable,
          heading.variable,
        )}
      >
        <Providers>{children}</Providers>
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
      <span className="text-icon" />
    </html>
  );
}
