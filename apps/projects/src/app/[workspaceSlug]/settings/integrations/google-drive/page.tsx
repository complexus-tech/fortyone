import { redirect } from "next/navigation";
import { withWorkspacePath } from "@/utils/workspace-url";

export const metadata = { title: "Settings › Google Drive" };

export default async function LegacyGoogleDriveIntegrationPage({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  redirect(withWorkspacePath("/settings/account/google-drive", workspaceSlug));
}
