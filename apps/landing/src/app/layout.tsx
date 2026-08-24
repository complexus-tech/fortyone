import "../styles/global.css";
import type { Metadata } from "next";
import { type ReactNode } from "react";
import { GoogleAnalytics, GoogleTagManager } from "@next/third-parties/google";
import { cn } from "lib";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  HOME_METADATA_DESCRIPTION,
  HOME_METADATA_TITLE,
} from "@/lib/seo";
import { body, mono, sans, serif } from "@/styles/fonts";
import { Toaster } from "./toaster";
import Providers from "./providers";

export const metadata: Metadata = {
  title: HOME_METADATA_TITLE,
  description: HOME_METADATA_DESCRIPTION,
  metadataBase: new URL("https://www.fortyone.app"),
  openGraph: {
    type: "website",
    locale: "en_US",
    title: HOME_METADATA_TITLE,
    description: HOME_METADATA_DESCRIPTION,
    images: [DEFAULT_SOCIAL_IMAGE],
    siteName: "FortyOne",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@fortyoneapp",
    creator: "@fortyoneapp",
    title: HOME_METADATA_TITLE,
    description: HOME_METADATA_DESCRIPTION,
    images: [DEFAULT_TWITTER_IMAGE],
  },
};
const isProduction = process.env.NODE_ENV === "production";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      className={cn(
        body.variable,
        sans.variable,
        mono.variable,
        serif.variable,
      )}
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
