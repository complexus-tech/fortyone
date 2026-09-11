import { notFound, permanentRedirect } from "next/navigation";
import { getIntegrationBySlug, integrations } from "@/lib/integrations";

export function generateStaticParams() {
  return integrations.map(({ slug }) => ({ slug }));
}

export default async function IntegrationPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  if (!getIntegrationBySlug(slug)) notFound();

  permanentRedirect(`/features/integrations#${slug}`);
}
