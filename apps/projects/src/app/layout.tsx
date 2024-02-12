import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import "../styles/global.css";
import { MainLayout } from "@/components/layout";

const font = Inter({
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Projects",
  description: "Complexus Projects",
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={font.className}>
        <MainLayout>{children}</MainLayout>
      </body>
    </html>
  );
}
