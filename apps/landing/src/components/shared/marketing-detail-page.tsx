import Image from "next/image";
import Link from "next/link";
import { CheckIcon, CloseIcon } from "icons";
import meshImage from "../../../public/images/meshing.webp";
import { CallToAction } from "./cta";

const UPDATED_DATE = "2026-06-24";
const UPDATED_LABEL = "June 24, 2026";
const cardTextClass = "text-[0.9rem] leading-[1.35]";
const cardMetaTextClass = "text-[0.82rem] leading-[1.25]";
const cardSurfaceClass =
  "bg-surface-elevated rounded-xl border border-border/80 shadow-lg shadow-shadow";

export type MarketingVisualRow = {
  label: string;
  value: string;
  width?: string;
};

export type MarketingVisual = {
  badge?: string;
  heading: string;
  note?: string;
  rows: MarketingVisualRow[];
  subheading: string;
};

export type MarketingSection = {
  cards?: MarketingVisual[];
  comparisonTable?: {
    competitor: string;
    rows: {
      competitor: boolean;
      feature: string;
      fortyOne: boolean;
      note: string;
    }[];
  };
  id: string;
  paragraphs: string[];
  prompt?: string;
  promptTitle?: string;
  rows?: [string, string][];
  tableHead?: [string, string];
  title: string;
};

export type MarketingDetail = {
  benefits: [string, string][];
  heroTitle: string;
  intro: string[];
  label: string;
  metaDescription: string;
  metaTitle: string;
  previewCards: MarketingVisual[];
  questions: [string, string][];
  sections: MarketingSection[];
  slug: string;
};

function PromptCard({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-surface-muted my-10 rounded-2xl p-2">
      <div className="px-3 py-2">
        <p className="text-text-muted text-sm font-medium">{title}</p>
      </div>
      <div className="bg-background text-foreground rounded-xl px-5 py-4 text-[1rem] leading-7">
        {children}
      </div>
    </div>
  );
}

function FeatureCardFrame({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative flex min-h-[320px] items-end overflow-hidden rounded-2xl">
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

function CardGrid({ children }: { children: React.ReactNode }) {
  return <div className="my-10 grid gap-6 md:grid-cols-2">{children}</div>;
}

function MockupCard({ card }: { card: MarketingVisual }) {
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
                <div className="bg-surface-muted h-2 overflow-hidden rounded-full">
                  <div
                    className="bg-foreground h-full rounded-full"
                    style={{ width: row.width }}
                  />
                </div>
              </div>
            ) : (
              <div
                className="bg-surface-muted rounded-lg px-3 py-2"
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

function AvailabilityMark({ available }: { available: boolean }) {
  return (
    <span
      aria-label={available ? "Available" : "Not available"}
      className={
        available
          ? "bg-background-inverse text-foreground-inverse inline-flex h-5 w-5 items-center justify-center rounded-full"
          : "bg-danger inline-flex h-5 w-5 items-center justify-center rounded-full text-white"
      }
    >
      {available ? (
        <CheckIcon className="h-3 w-auto" strokeWidth={2.4} />
      ) : (
        <CloseIcon className="h-3 w-auto text-white" strokeWidth={2.4} />
      )}
    </span>
  );
}

function ComparisonTable({
  competitor,
  rows,
}: NonNullable<MarketingSection["comparisonTable"]>) {
  return (
    <div className="border-border my-10 overflow-x-auto border-y">
      <table className="w-full min-w-[620px] text-left text-[0.95rem]">
        <thead>
          <tr className="border-border text-foreground border-b">
            <th className="w-[36%] py-4 pr-6 font-semibold">Capability</th>
            <th className="w-[18%] py-4 pr-6 text-center font-semibold">
              FortyOne
            </th>
            <th className="w-[18%] py-4 pr-6 text-center font-semibold">
              {competitor}
            </th>
            <th className="py-4 font-semibold">Planning impact</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              className="border-border border-b last:border-b-0"
              key={row.feature}
            >
              <td className="text-foreground py-4 pr-6 align-top font-medium">
                {row.feature}
              </td>
              <td className="py-4 pr-6 align-top">
                <div className="flex justify-center">
                  <AvailabilityMark available={row.fortyOne} />
                </div>
              </td>
              <td className="py-4 pr-6 align-top">
                <div className="flex justify-center">
                  <AvailabilityMark available={row.competitor} />
                </div>
              </td>
              <td className="text-text-muted py-4 leading-7">{row.note}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function JsonLd({
  basePath,
  canonicalPath,
  detail,
}: {
  basePath: string;
  canonicalPath?: string;
  detail: MarketingDetail;
}) {
  const path = canonicalPath ?? `${basePath}/${detail.slug}`;
  const url = `https://www.fortyone.app/${path}`;
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    name: detail.heroTitle,
    description: detail.metaDescription,
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

export function MarketingDetailPage({
  basePath,
  breadcrumbLabel,
  canonicalPath,
  detail,
  questionHeading,
}: {
  basePath: string;
  breadcrumbLabel: string;
  canonicalPath?: string;
  detail: MarketingDetail;
  questionHeading?: string;
}) {
  return (
    <>
      <JsonLd
        basePath={basePath}
        canonicalPath={canonicalPath}
        detail={detail}
      />
      <main className="bg-background text-foreground">
        <div className="mx-auto max-w-[760px] px-5 pt-20 pb-16 md:px-8 md:pt-28">
          <article>
            <nav className="text-text-muted mb-6 flex justify-center gap-2 text-sm">
              <Link className="hover:text-foreground" href="/">
                Home
              </Link>
              <span>/</span>
              <span>{breadcrumbLabel}</span>
            </nav>

            <header className="text-center">
              <div className="shadow-shadow relative mx-auto max-w-[720px] overflow-hidden rounded-2xl shadow-2xl">
                <Image
                  alt=""
                  className="h-[240px] w-full object-cover md:h-[320px]"
                  placeholder="blur"
                  priority
                  src={meshImage}
                />
                <div className="absolute inset-0 bg-black/20 dark:bg-black/35" />
                <div className="absolute inset-0 flex items-center justify-center px-6">
                  <h1 className="font-heading text-4xl font-semibold tracking-normal text-white md:text-5xl">
                    {detail.heroTitle}
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
              {detail.intro.map((paragraph) => (
                <p className="mb-6" key={paragraph}>
                  {paragraph}
                </p>
              ))}

              <CardGrid>
                {detail.previewCards.map((card) => (
                  <FeatureCardFrame key={card.heading}>
                    <MockupCard card={card} />
                  </FeatureCardFrame>
                ))}
              </CardGrid>

              <section className="scroll-mt-28" id="benefits">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Where FortyOne helps
                </h2>
                <DataTable
                  rows={detail.benefits}
                  tableHead={["Benefit", "Why it matters"]}
                />
              </section>

              {detail.sections.map((section) => (
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
                    <CardGrid>
                      {section.cards.map((card) => (
                        <FeatureCardFrame key={card.heading}>
                          <MockupCard card={card} />
                        </FeatureCardFrame>
                      ))}
                    </CardGrid>
                  ) : null}
                  {section.rows ? (
                    <DataTable
                      rows={section.rows}
                      tableHead={section.tableHead}
                    />
                  ) : null}
                  {section.comparisonTable ? (
                    <ComparisonTable {...section.comparisonTable} />
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
                  {questionHeading ??
                    `Questions about ${detail.label.toLowerCase()}`}
                </h2>
                <div className="divide-border border-border divide-y border-y">
                  {detail.questions.map(([question, answer]) => (
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
