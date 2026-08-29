import type { Metadata } from "next";
import { SlackAccountLinkSettings } from "@/modules/settings/workspace/integrations/slack";

export const metadata: Metadata = {
  title: "Settings › Slack",
};

export default function SlackAccountLinkPage() {
  return <SlackAccountLinkSettings />;
}
