import type { Metadata } from "next";
import type { ReactNode } from "react";
import { SettingsLayout } from "@/components/layouts";

export const metadata: Metadata = {
  title: "Settings",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return <SettingsLayout>{children}</SettingsLayout>;
}
