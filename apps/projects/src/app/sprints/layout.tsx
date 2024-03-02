import type { Metadata } from "next";
import type { ReactNode } from "react";
import { MainLayout } from "@/components/layout";

export const metadata: Metadata = {
  title: "Projects",
  description: "Complexus Projects",
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return <MainLayout>{children}</MainLayout>;
}
