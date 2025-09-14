import type { Metadata } from "next";
import { Billing } from "@/modules/settings/workspace/billing";

export const metadata: Metadata = {
  title: "Settings › Billing",
};

export default function ObjectivesWorkflowPage() {
  return <Billing />;
}
