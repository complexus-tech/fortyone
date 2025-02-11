import type { Metadata } from "next";
import type { ReactNode } from "react";
import { cn } from "lib";
import { instrumentSans, satoshi } from "@/styles/fonts";
import { CallToAction, Footer, Navigation } from "@/components/shared";
import "../styles/global.css";
import { CursorProvider } from "@/context";
import { PostHogProvider } from "./posthog";
// import dynamic from "next/dynamic";

// const PostHogPageView = dynamic(() => import("./posthog-page-view"), {
//   ssr: false,
// });

export const metadata: Metadata = {
  title: "Nail every objective on time with complexus.",
  description:
    "Empower your team to crush every key objective with our seamless project management platform.",
  openGraph: {
    title: "Nail every objective on time with complexus.",
    description:
      "Empower your team to crush every key objective with our seamless project management platform.",
    url: "https://complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_tech",
    title: "Nail every objective on time with complexus.",
    description:
      "Empower your team to crush every key objective with our seamless project management platform.",
  },
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return (
    <html className="dark" lang="en" suppressHydrationWarning>
      <body className={cn(satoshi.variable, instrumentSans.variable)}>
        <PostHogProvider>
          <CursorProvider>
            <Navigation />
            {children}
            <CallToAction />
            <Footer />
          </CursorProvider>
          {/* <PostHogPageView /> */}
        </PostHogProvider>
      </body>
    </html>
  );
}
