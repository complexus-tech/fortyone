import type { Metadata } from "next";
import { DocumentPage } from "@/modules/documents/document-page";

export const metadata: Metadata = { title: "Document" };

export default async function Page({
  params,
}: {
  params: Promise<{ documentId: string }>;
}) {
  const { documentId } = await params;
  return <DocumentPage documentId={documentId} />;
}
