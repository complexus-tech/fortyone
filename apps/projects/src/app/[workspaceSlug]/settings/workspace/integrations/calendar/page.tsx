import { redirect } from "next/navigation";
import { withWorkspacePath } from "@/utils";

export const metadata = {
  title: "Settings › Calendar",
};

export default async function CalendarIntegrationPage({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  redirect(withWorkspacePath("/settings/account/calendar", workspaceSlug));
}
