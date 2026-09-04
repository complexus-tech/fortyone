import { redirect } from "next/navigation";
import { withWorkspacePath } from "@/utils";

export const metadata = { title: "Settings › Google Drive" };

export default async function LegacyGoogleDriveIntegrationPage({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  redirect(withWorkspacePath("/settings/account/google-drive", workspaceSlug));
}
