import type { Metadata } from "next";
import { WorkspaceSecuritySettings } from "@/modules/settings/workspace/security";

export const metadata: Metadata = { title: "Settings › Workspace security" };
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string }>;
}) {
  const { tab } = await searchParams;
  const initialTab = [
    "policies",
    "sso",
    "sessions",
    "audit",
    "provisioning",
  ].includes(tab ?? "")
    ? tab
    : "policies";
  return <WorkspaceSecuritySettings initialTab={initialTab} />;
}
