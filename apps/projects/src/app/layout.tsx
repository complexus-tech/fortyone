import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import { Toaster } from "sonner";
import "../styles/global.css";
import { ProgressBar } from "./progress";
import { Providers } from "./providers";
import dynamic from "next/dynamic";
import { toasterIcons } from "./toaster-icons";

const PostHogPageView = dynamic(() => import("./posthog-page-view"), {
  ssr: false,
});

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
      <body className={font.className}>
        <Providers>{children}</Providers>
        <Toaster
          theme="system"
          closeButton
          position="bottom-right"
          duration={10000}
          toastOptions={{
            className:
              "w-full rounded-lg p-4 flex items-center gap-3 shadow-lg",
            classNames: {
              toast:
                "bg-white/90 dark:bg-dark-100/90 backdrop-blur border border-gray-100/60 dark:border-dark-50",
              closeButton:
                "bg-white/90 dark:bg-dark-100/90 dark:border-dark-50",
            },
          }}
          icons={toasterIcons}
        />
        <PostHogPageView />
        <ProgressBar />
      </body>
    </html>
  );
}
