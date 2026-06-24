import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { CallToAction } from "@/components/shared";
import {
  getUseCaseBySlug,
  useCases,
  type UseCaseVisual,
} from "@/lib/use-cases";
import meshImage from "../../../../../public/images/meshing.webp";

const UPDATED_DATE = "2026-06-24";
const UPDATED_LABEL = "June 24, 2026";
const cardTextClass = "text-[0.9rem] leading-[1.35]";
const cardMetaTextClass = "text-[0.82rem] leading-[1.25]";
const cardSurfaceClass =
  "rounded-xl border border-white/50 bg-background shadow-lg shadow-shadow dark:border-border";

export function generateStaticParams() {
  return useCases.map((useCase) => ({ slug: useCase.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const useCase = getUseCaseBySlug(slug);

  if (!useCase) {
    return {};
  }

  const canonicalUrl = `https://www.fortyone.app/use-cases/${useCase.slug}`;

  return {
    title: useCase.metaTitle,
    description: useCase.metaDescription,
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      type: "website",
      url: canonicalUrl,
      title: useCase.metaTitle,
      description: useCase.metaDescription,
      siteName: "FortyOne",
    },
    twitter: {
      card: "summary_large_image",
      title: useCase.metaTitle,
      description: useCase.metaDescription,
    },
  };
}

function PromptCard({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="my-10 rounded-2xl bg-black/[0.07] p-2 dark:bg-white/[0.12]">
      <div className="px-3 py-2">
        <p className="text-text-muted text-sm font-medium">{title}</p>
      </div>
      <div className="bg-background text-foreground rounded-xl px-5 py-4 text-[1rem] leading-7">
        {children}
      </div>
    </div>
  );
}

function UseCaseFeatureCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative flex aspect-[4/3] items-end overflow-hidden rounded-2xl">
      <Image
        alt=""
        className="object-cover dark:opacity-40"
        fill
        quality={100}
        sizes="(max-width: 767px) 100vw, 380px"
        src={meshImage}
      />
      <div className="relative z-10 w-full p-4">{children}</div>
    </div>
  );
}

function UseCaseCardGrid({ children }: { children: React.ReactNode }) {
  return <div className="my-10 grid gap-6 md:grid-cols-2">{children}</div>;
}

function UseCaseMockupCard({ card }: { card: UseCaseVisual }) {
  return (
    <div className="flex h-full flex-col gap-2.5">
      <div className={`${cardSurfaceClass} px-3 py-2.5`}>
        <div className="flex items-center justify-between gap-3">
          <div className="min-w-0">
            <p className={`${cardTextClass} text-foreground font-semibold`}>
              {card.heading}
            </p>
            <p className={`${cardMetaTextClass} text-text-muted truncate`}>
              {card.subheading}
            </p>
          </div>
          {card.badge ? (
            <span className="bg-success/10 text-success shrink-0 rounded-lg px-2 py-1 text-xs font-semibold">
              {card.badge}
            </span>
          ) : null}
        </div>
      </div>
      <div className={`${cardSurfaceClass} flex-1 p-3`}>
        <div className="grid gap-2">
          {card.rows.map((row) =>
            row.width ? (
              <div key={row.label}>
                <div className="mb-2 flex items-center justify-between">
                  <span
                    className={`${cardMetaTextClass} text-foreground font-medium`}
                  >
                    {row.label}
                  </span>
                  <span className={`${cardMetaTextClass} text-text-muted`}>
                    {row.value}
                  </span>
                </div>
                <div className="bg-surface-muted h-2 overflow-hidden rounded-full dark:bg-white/10">
                  <div
                    className="bg-foreground h-full rounded-full"
                    style={{ width: row.width }}
                  />
                </div>
              </div>
            ) : (
              <div
                className="rounded-lg bg-black/4 px-3 py-2 dark:bg-white/7"
                key={row.label}
              >
                <p
                  className={`${cardMetaTextClass} text-foreground font-semibold`}
                >
                  {row.label}
                </p>
                <p className={`${cardMetaTextClass} text-text-muted mt-1`}>
                  {row.value}
                </p>
              </div>
            ),
          )}
        </div>
        {card.note ? (
          <p
            className={`${cardMetaTextClass} bg-accent text-text-secondary mt-3 rounded-lg px-3 py-2 font-medium`}
          >
            {card.note}
          </p>
        ) : null}
      </div>
    </div>
  );
}

function DataTable({
  rows,
  tableHead = ["Input", "What FortyOne prepares"],
}: {
  rows: [string, string][];
  tableHead?: [string, string];
}) {
  return (
    <div className="border-border my-10 overflow-hidden border-y">
      <table className="w-full text-left text-[0.95rem]">
        <thead>
          <tr className="border-border text-foreground border-b">
            <th className="w-[34%] py-4 pr-6 font-semibold">{tableHead[0]}</th>
            <th className="py-4 font-semibold">{tableHead[1]}</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(([label, value]) => (
            <tr className="border-border border-b last:border-b-0" key={label}>
              <td className="text-foreground py-4 pr-6 align-top font-medium">
                {label}
              </td>
              <td className="text-text-muted py-4 leading-7">{value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function JsonLd({
  slug,
  title,
  description,
}: {
  description: string;
  slug: string;
  title: string;
}) {
  const url = `https://www.fortyone.app/use-cases/${slug}`;
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    name: title,
    description,
    url,
    publisher: {
      "@type": "Organization",
      name: "FortyOne",
    },
  };

  return (
    <script
      dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      type="application/ld+json"
    />
  );
}

export default async function UseCasePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const useCase = getUseCaseBySlug(slug);

  if (!useCase) {
    return notFound();
  }

  return (
    <>
      <JsonLd
        description={useCase.metaDescription}
        slug={useCase.slug}
        title={useCase.heroTitle}
      />
      <main className="bg-background text-foreground">
        <div className="mx-auto max-w-[760px] px-5 pt-20 pb-16 md:px-8 md:pt-28">
          <article>
            <nav className="text-text-muted mb-6 flex justify-center gap-2 text-sm">
              <Link className="hover:text-foreground" href="/">
                Home
              </Link>
              <span>/</span>
              <span>Use cases</span>
            </nav>

            <header className="text-center">
              <div className="shadow-shadow relative mx-auto max-w-[720px] overflow-hidden rounded-2xl shadow-2xl">
                <Image
                  alt=""
                  className="h-[240px] w-full object-cover md:h-[320px]"
                  priority
                  placeholder="blur"
                  src={meshImage}
                />
                <div className="absolute inset-0 bg-black/20 dark:bg-black/35" />
                <div className="absolute inset-0 flex items-center justify-center px-6">
                  <h1 className="font-heading text-4xl font-semibold tracking-normal text-white md:text-5xl">
                    {useCase.heroTitle}
                  </h1>
                </div>
              </div>
              <time
                className="text-text-muted mt-5 block text-sm font-medium"
                dateTime={UPDATED_DATE}
              >
                {UPDATED_LABEL}
              </time>
            </header>

            <div className="text-text-muted mt-18 text-[1.05rem] leading-8">
              {useCase.intro.map((paragraph) => (
                <p className="mb-6" key={paragraph}>
                  {paragraph}
                </p>
              ))}

              <UseCaseCardGrid>
                {useCase.previewCards.map((card) => (
                  <UseCaseFeatureCard key={card.heading}>
                    <UseCaseMockupCard card={card} />
                  </UseCaseFeatureCard>
                ))}
              </UseCaseCardGrid>

              <section className="scroll-mt-28" id="benefits">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Where FortyOne helps
                </h2>
                <DataTable
                  rows={useCase.benefits}
                  tableHead={["Benefit", "Why it matters"]}
                />
              </section>

              {useCase.sections.map((section) => (
                <section
                  className="scroll-mt-28"
                  id={section.id}
                  key={section.id}
                >
                  <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                    {section.title}
                  </h2>
                  {section.paragraphs.map((paragraph) => (
                    <p className="mb-6" key={paragraph}>
                      {paragraph}
                    </p>
                  ))}
                  {section.cards ? (
                    <UseCaseCardGrid>
                      {section.cards.map((card) => (
                        <UseCaseFeatureCard key={card.heading}>
                          <UseCaseMockupCard card={card} />
                        </UseCaseFeatureCard>
                      ))}
                    </UseCaseCardGrid>
                  ) : null}
                  {section.rows ? (
                    <DataTable
                      rows={section.rows}
                      tableHead={section.tableHead}
                    />
                  ) : null}
                  {section.prompt && section.promptTitle ? (
                    <PromptCard title={section.promptTitle}>
                      {section.prompt}
                    </PromptCard>
                  ) : null}
                </section>
              ))}

              <section className="scroll-mt-28" id="questions">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Questions from {useCase.label.toLowerCase()}
                </h2>
                <div className="divide-border border-border divide-y border-y">
                  {useCase.questions.map(([question, answer]) => (
                    <div className="py-5" key={question}>
                      <h3 className="text-foreground mb-2 text-xl font-semibold">
                        {question}
                      </h3>
                      <p>{answer}</p>
                    </div>
                  ))}
                </div>
              </section>
            </div>
          </article>
        </div>
      </main>
      <CallToAction />
    </>
  );
}
