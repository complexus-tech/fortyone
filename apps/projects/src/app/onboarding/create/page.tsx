import type { Metadata } from "next";
import { CreateWorkspace } from "@/modules/onboarding/create";

export const metadata: Metadata = {
  title: "Create Workspace - FortyOne",
  description: "Create a new workspace",
};

export default async function CreateWorkspacePage({
  searchParams,
}: {
  searchParams: Promise<{ callbackUrl?: string }>;
}) {
  const { callbackUrl } = await searchParams;

  return <CreateWorkspace callbackUrl={callbackUrl} />;
}
