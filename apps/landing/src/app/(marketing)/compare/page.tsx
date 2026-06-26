import type { Metadata } from "next";
import Link from "next/link";
import { CallToAction } from "@/components/shared/cta";
import { Container } from "@/components/ui";
import { comparisons } from "@/lib/comparisons";

export const metadata: Metadata = {
  title: "Compare FortyOne | AI Project Management Alternatives",
  description:
    "Compare FortyOne with Linear, Asana, Jira, ClickUp, and monday.com for AI planning, goals, tasks, roadmaps, and review controls.",
  alternates: {
    canonical: "https://www.fortyone.app/compare",
  },
};

export default function ComparePage() {
  return (
    <>
      <main className="bg-background text-foreground">
        <Container className="pt-28 pb-16 md:pt-36 md:pb-24">
          <div className="max-w-3xl">
            <p className="text-text-muted mb-5 text-sm font-semibold">
              Compare
            </p>
            <h1 className="font-heading text-5xl leading-[0.98] font-semibold tracking-normal md:text-7xl">
              Pick the project system that fits how your team plans.
            </h1>
            <p className="text-text-muted mt-6 max-w-2xl text-lg leading-8">
              FortyOne is different when teams need goal-connected tasks, AI
              owner and estimate suggestions, roadmap execution, and review
              controls before AI changes apply.
            </p>
          </div>
        </Container>

        <Container className="pb-20 md:pb-28">
          <div className="grid gap-4 md:grid-cols-2">
            {comparisons.map((comparison) => (
              <Link
                className="border-border bg-surface hover:bg-state-hover group rounded-3xl border p-5 transition-colors"
                href={`/compare/${comparison.slug}`}
                key={comparison.slug}
              >
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-text-muted text-sm font-semibold">
                      FortyOne vs
                    </p>
                    <h2 className="mt-2 text-3xl font-semibold tracking-normal">
                      {comparison.competitor}
                    </h2>
                  </div>
                  <span className="border-border group-hover:border-foreground rounded-full border px-3 py-1 text-sm transition">
                    Compare
                  </span>
                </div>
                <p className="text-text-muted mt-5 leading-7">
                  {comparison.summary}
                </p>
              </Link>
            ))}
          </div>
        </Container>
      </main>
      <CallToAction />
    </>
  );
}
