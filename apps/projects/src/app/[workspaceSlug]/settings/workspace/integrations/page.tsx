import { redirect } from "next/navigation";
import { withWorkspacePath } from "@/utils";

export const metadata = {
  title: "Settings › Integrations",
};

export default async function LegacyWorkspaceIntegrationsPage({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  redirect(withWorkspacePath("/settings/integrations", workspaceSlug));
}
