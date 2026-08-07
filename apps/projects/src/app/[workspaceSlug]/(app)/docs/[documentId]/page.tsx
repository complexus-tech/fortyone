import type { Metadata } from "next";
import { DocumentPage } from "@/modules/documents/document-page";
import { auth } from "@/auth";
import { getDocument } from "@/modules/documents/queries";

type DocumentRouteProps = {
  params: Promise<{ documentId: string; workspaceSlug: string }>;
};

export async function generateMetadata({
  params,
}: DocumentRouteProps): Promise<Metadata> {
  const [{ documentId, workspaceSlug }, session] = await Promise.all([
    params,
    auth(),
  ]);
  const document = await getDocument(documentId, {
    session: session!,
    workspaceSlug,
  });

  return {
    title: document.title.trim() || "Untitled document",
  };
}

export default async function Page({ params }: DocumentRouteProps) {
  const { documentId } = await params;
  return <DocumentPage documentId={documentId} />;
}
