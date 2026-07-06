import type { Metadata } from "next";
import { SlackIntegrationSettings } from "@/modules/settings/workspace/integrations/slack";

export const metadata: Metadata = {
  title: "Settings › Slack",
};

export default function SlackIntegrationPage() {
  return <SlackIntegrationSettings />;
}
