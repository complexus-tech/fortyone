import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { PublicBranding } from "@/components/ui/public-branding";
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
      <article className="mx-auto max-w-3xl px-6 pt-14 pb-28 md:pt-20">
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
      <PublicBranding />
    </main>
  );
}
