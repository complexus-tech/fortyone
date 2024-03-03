import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import "../styles/global.css";

const font = Inter({
  subsets: ["latin"],
  display: "swap",
  weight: ["500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Complexus: The Collaborative Hub for Seamless Project Management",
  description:
    "Simplify workflows, boost collaboration, and deliver projects on time with Complexus. Our intuitive platform empowers teams to visualize progress, gain data-driven insights, and achieve success.",
  openGraph: {
    title: "Complexus: Orchestrate Your Projects to Success",
    description:
      "Simplify workflows, boost collaboration, and deliver on time with Complexus. We offer a collaborative hub for seamless project management.",
    url: "https://complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_tech",
    title: "Complexus: Simplify & Succeed in Project Management",
    description:
      "Effortlessly manage projects with Complexus. Boost team collaboration, track progress, and make data-driven decisions.",
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
    </html>
  );
}
