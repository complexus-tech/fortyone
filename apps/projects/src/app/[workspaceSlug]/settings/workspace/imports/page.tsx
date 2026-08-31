import type { Metadata } from "next";
import { WorkspaceImportSettings } from "@/modules/settings/workspace/imports";

export const metadata: Metadata = {
  title: "Settings › Imports",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ from?: string | string[] }>;
}) {
  const query = await searchParams;
  return (
    <WorkspaceImportSettings openFromOnboarding={query.from === "onboarding"} />
  );
}
