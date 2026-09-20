import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { PublicDocumentLayout } from "@/modules/documents/public-document-layout";
import { getApiUrl } from "@/lib/api-url";
import { richTextHTMLToMarkdown } from "@/lib/tiptap/markdown-server";
import { publicDocumentHTML } from "@/modules/documents/public-document-html";

export const metadata: Metadata = {
  title: "Shared document · FortyOne",
  robots: { index: false, follow: false },
  referrer: "no-referrer",
};

export default async function SharedDocumentPage({
  params,
}: {
  params: Promise<{ token: string }>;
}) {
  const { token } = await params;
  if (!/^[0-9a-f]{64}$/.test(token)) notFound();
  const apiUrl = getApiUrl();
  const response = await fetch(`${apiUrl}/public/documents/${token}`, {
    cache: "no-store",
  });
  if (response.status === 404) notFound();
  if (!response.ok)
    throw new Error("The shared document is temporarily unavailable");
  const { data } = (await response.json()) as {
    data: {
      title: string;
      contentHtml: string;
      contentText: string;
      updatedAt: string;
    };
  };
  const contentHtml = publicDocumentHTML(data.contentHtml, token, apiUrl);
  const updatedAt = new Date(data.updatedAt).toLocaleDateString("en", {
    dateStyle: "long",
  });
  return (
    <PublicDocumentLayout
      contentHtml={contentHtml}
      markdown={richTextHTMLToMarkdown(contentHtml)}
      title={data.title}
      updatedAt={updatedAt}
    />
  );
}
