import type { Metadata } from "next";
import { GitHubIntegrationSettings } from "@/modules/settings/workspace/integrations/github";

export const metadata: Metadata = {
  title: "Settings › GitHub",
};

export default function GitHubIntegrationPage() {
  return <GitHubIntegrationSettings />;
}
