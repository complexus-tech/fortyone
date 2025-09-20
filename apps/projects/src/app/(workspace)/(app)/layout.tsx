import type { ReactNode } from "react";
import { ApplicationLayout } from "@/components/layouts";

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  return <ApplicationLayout>{children}</ApplicationLayout>;
}
