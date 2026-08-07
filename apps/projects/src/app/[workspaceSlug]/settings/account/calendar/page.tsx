import { redirect } from "next/navigation";
import { withWorkspacePath } from "@/utils";

export const metadata = { title: "Settings › Google Calendar" };

export default async function LegacyAccountCalendarPage({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  redirect(withWorkspacePath("/settings/integrations/calendar", workspaceSlug));
}
