import "./global.css";
import { RootProvider } from "fumadocs-ui/provider/next";
import { cn } from "lib";
import {
  Bricolage_Grotesque as BricolageGrotesque,
  Geist,
} from "next/font/google";
import type { Metadata } from "next";
import Script from "next/script";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  metadataBase: new URL("https://docs.fortyone.app"),
};

const geist = Geist({
  variable: "--font-geist",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

const heading = BricolageGrotesque({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className={cn(geist.variable, heading.variable)}
      suppressHydrationWarning
    >
      <body className="flex flex-col min-h-screen antialiased">
        <RootProvider>{children}</RootProvider>
        <Script
          async
          data-default-tab="home"
          data-mode="bubble"
          data-portal="complexus"
          data-position="bottom-right"
          data-theme="auto"
          src="https://complexus.fortyone.app/api/feedback-widget/v1.js"
          strategy="afterInteractive"
        />
      </body>
    </html>
  );
}
