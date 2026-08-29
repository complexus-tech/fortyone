import type { Metadata } from "next";
import { IntegrationRequestDetails } from "@/modules/integration-requests/details";

export const metadata: Metadata = {
  title: "Intake",
};

export default async function Page({
  params,
}: {
  params: Promise<{ requestId: string }>;
}) {
  const { requestId } = await params;
  return <IntegrationRequestDetails requestId={requestId} />;
}
