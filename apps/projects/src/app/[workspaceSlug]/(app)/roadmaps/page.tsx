import { redirect } from "next/navigation";

export default async function Page({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  redirect(`/${workspaceSlug}/objectives`);
}
