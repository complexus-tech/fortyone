import type { Metadata } from "next";
import type { ReactNode } from "react";
import { ApplicationLayout } from "@/components/layouts";

export const metadata: Metadata = {
  title: "Objectives",
  description: "Complexus Objectives",
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return <ApplicationLayout>{children}</ApplicationLayout>;
}
