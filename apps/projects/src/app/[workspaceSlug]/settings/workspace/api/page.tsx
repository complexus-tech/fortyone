import type { Metadata } from "next";
import { ApiSettings } from "@/modules/settings/workspace/api";

export const metadata: Metadata = {
  title: "Settings › API",
};

export default function Page() {
  return <ApiSettings />;
}
