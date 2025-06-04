import type { Metadata } from "next";
import { AnalyticsPage } from "@/modules/analytics";

export const metadata: Metadata = {
  title: "Reports",
};

export default function Page() {
  return <AnalyticsPage />;
}
