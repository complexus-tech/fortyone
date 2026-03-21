import type { Metadata } from "next";
import { IntegrationsIndex } from "@/modules/settings/workspace/integrations";

export const metadata: Metadata = {
  title: "Settings › Integrations",
};

export default function IntegrationsPage() {
  return <IntegrationsIndex />;
}
