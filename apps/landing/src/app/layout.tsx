import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import { CallToAction, Footer, Navigation } from "@/components/shared";
import "../styles/global.css";
import { CursorProvider } from "@/context";

const font = Inter({
  subsets: ["latin"],
  display: "swap",
  weight: ["300", "400", "500", "600", "700"],
});

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
      <body className={font.className}>
        <CursorProvider>
          <Navigation />
          {children}
          <CallToAction />
          <Footer />
        </CursorProvider>
      </body>
    </html>
  );
}
