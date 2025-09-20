import type { Metadata } from "next";
import { SecuritySettings } from "@/modules/settings/account/security";

export const metadata: Metadata = {
  title: "Settings › Security",
};

export default function Page() {
  return <SecuritySettings />;
}
