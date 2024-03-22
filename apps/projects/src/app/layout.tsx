import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import "../styles/global.css";
import { ProgressBar } from "./progress";

const font = Inter({
  subsets: ["latin"],
  display: "swap",
  weight: ["500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Complexus: The Collaborative Hub for Seamless Objective Management",
  description:
    "Simplify workflows, boost collaboration, and deliver objectives on time with Complexus. Our intuitive platform empowers teams to visualize progress, gain data-driven insights, and achieve success.",
  openGraph: {
    title: "Complexus: Orchestrate Your Objectives to Success",
    description:
      "Simplify workflows, boost collaboration, and deliver on time with Complexus. We offer a collaborative hub for seamless objective management.",
    url: "https://complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_tech",
    title: "Complexus: Simplify & Succeed in Objective Management",
    description:
      "Effortlessly manage objectives with Complexus. Boost team collaboration, track progress, and make data-driven decisions.",
  },
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={font.className}>{children}</body>
      <ProgressBar />
    </html>
  );
}
