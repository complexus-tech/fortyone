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
  title:
    "Complexus: Empowering Teams to Conquer Project Complexity, Effortlessly.",
  description:
    "Empower your team with innovative project management solutions. Simplify workflows, streamline collaboration, and achieve exceptional results with Complexus. With features like OKR Tracking, Epics Management, Iterations Planning, and Roadmap Visualization, welcome to effortless project management.",
  openGraph: {
    title:
      "Complexus: Empowering Teams to Conquer Project Complexity, Effortlessly.",
    description:
      "Empower your team with innovative project management solutions. Simplify workflows, streamline collaboration, and achieve exceptional results with Complexus. With features like OKR Tracking, Epics Management, Iterations Planning, and Roadmap Visualization, welcome to effortless project management.",
    url: "https://complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_tech",
    title:
      "Complexus: Empowering Teams to Conquer Project Complexity, Effortlessly.",
    description:
      "Empower your team with innovative project management solutions. Simplify workflows, streamline collaboration, and achieve exceptional results with Complexus. With features like OKR Tracking, Epics Management, Iterations Planning, and Roadmap Visualization, welcome to effortless project management.",
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
