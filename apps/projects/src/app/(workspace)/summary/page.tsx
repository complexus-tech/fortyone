import type { Metadata } from "next";
import { SummaryPage } from "@/modules/summary";

export const metadata: Metadata = {
  title: "Summary",
};

export default function Page() {
  return <SummaryPage />;
}
