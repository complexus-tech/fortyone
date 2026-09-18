import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getApiUrl } from "@/lib/api-url";
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
    data: { title: string; contentHtml: string; updatedAt: string };
  };
  return (
    <main className="bg-background text-foreground min-h-dvh">
      <header className="border-border border-b px-6 py-5">
        <div className="mx-auto flex max-w-4xl items-center justify-between">
          <span className="font-semibold">FortyOne</span>
          <span className="text-text-muted">Shared document · View only</span>
        </div>
      </header>
      <article className="mx-auto max-w-4xl px-6 py-14 md:py-20">
        <h1 className="mb-4 text-4xl leading-tight font-semibold md:text-5xl">
          {data.title}
        </h1>
        <p className="text-text-muted mb-10">
          Updated{" "}
          {new Date(data.updatedAt).toLocaleDateString("en", {
            dateStyle: "long",
          })}
        </p>
        <div
          className="rich-document-editor prose prose-lg dark:prose-invert prose-img:max-w-full max-w-none"
          dangerouslySetInnerHTML={{
            __html: publicDocumentHTML(data.contentHtml, token, apiUrl),
          }}
        />
      </article>
    </main>
  );
}
