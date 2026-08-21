import type { Metadata } from "next";
import { FigmaIntegrationSettings } from "@/modules/settings/workspace/integrations/figma";

export const metadata: Metadata = { title: "Settings › Figma" };

export default function FigmaIntegrationPage() {
  return <FigmaIntegrationSettings />;
}
