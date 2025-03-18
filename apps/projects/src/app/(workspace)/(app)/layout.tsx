import type { ReactNode } from "react";
import { ApplicationLayout } from "@/components/layouts";

export default function RootLayout({ children }: { children: ReactNode }) {
  return <ApplicationLayout>{children}</ApplicationLayout>;
}
