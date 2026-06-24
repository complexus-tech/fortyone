import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { CallToAction } from "@/components/shared";
import meshImage from "../../../../../public/images/meshing.webp";

const packetRows: Array<[string, string]> = [
  ["Project status", "At-risk tasks, blocked owners, and work due this week"],
  ["Capacity", "Which teams can take new work and which teams are near limit"],
  ["Assignments", "Suggested owners, estimates, and schedule windows"],
  ["Review note", "What changed, what needs approval, and what can wait"],
];

const intakeRows: Array<[string, string]> = [
  ["Slack request", "Create a task, attach context, and suggest an owner"],
  ["Calendar signal", "Find a realistic work window before assignment"],
  ["Missing estimate", "Fill an initial estimate from similar prior work"],
  ["Manager review", "Stop before applying important AI changes"],
];

const cardTextClass = "text-[0.9rem] leading-[1.35]";
const cardMetaTextClass = "text-[0.82rem] leading-[1.25]";
const cardSurfaceClass =
  "rounded-xl border border-white/50 bg-background shadow-lg shadow-shadow dark:border-border";

export const metadata: Metadata = {
  title: "FortyOne for Operations | AI Project Management Use Case",
  description:
    "See how operations teams use FortyOne to turn requests, capacity, assignments, and project reviews into one AI-assisted project management workflow.",
};

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

function OperationsPacketCard() {
  return (
    <div className="flex h-full flex-col gap-2.5">
      <div className={`${cardSurfaceClass} px-3 py-2.5`}>
        <div className="flex items-center justify-between gap-3">
          <div>
            <p className={`${cardTextClass} text-foreground font-semibold`}>
              Weekly operations review
            </p>
            <p className={`${cardMetaTextClass} text-text-muted`}>
              Prepared from live projects
            </p>
          </div>
          <span className="bg-success/10 text-success rounded-lg px-2 py-1 text-xs font-semibold">
            Ready
          </span>
        </div>
      </div>
      <div className={`${cardSurfaceClass} flex-1 p-3`}>
        <div className="grid gap-2">
          {[
            ["Blocked work", "4 tasks need owner decisions"],
            ["Capacity", "Product can take one more request"],
          ].map(([label, value]) => (
            <div
              className="rounded-lg bg-black/4 px-3 py-2 dark:bg-white/7"
              key={label}
            >
              <p
                className={`${cardMetaTextClass} text-foreground font-semibold`}
              >
                {label}
              </p>
              <p className={`${cardMetaTextClass} text-text-muted mt-1`}>
                {value}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function TaskIntakeCard() {
  return (
    <div className="flex h-full flex-col gap-2.5">
      <div className={`${cardSurfaceClass} px-3 py-2.5`}>
        <p className={`${cardTextClass} text-foreground font-semibold`}>
          Slack request
        </p>
        <p className={`${cardMetaTextClass} text-text-muted mt-1`}>
          &quot;Can we move the onboarding report into this week&apos;s ops
          review?&quot;
        </p>
      </div>
      <div className={`${cardSurfaceClass} flex-1 p-3`}>
        <div className="mb-2 flex items-center justify-between gap-3">
          <span className="bg-accent text-text-secondary rounded-lg px-2.5 py-1 text-xs font-semibold">
            AI draft
          </span>
          <span className={`${cardMetaTextClass} text-text-muted font-medium`}>
            Needs review
          </span>
        </div>
        {[
          ["Task", "Prepare onboarding report"],
          ["Owner", "Operations"],
        ].map(([label, value]) => (
          <div className="flex justify-between gap-4 py-2" key={label}>
            <span className={`${cardMetaTextClass} text-text-muted`}>
              {label}
            </span>
            <span
              className={`${cardMetaTextClass} text-foreground text-right font-semibold`}
            >
              {value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function CapacityPlanCard() {
  return (
    <div className="flex h-full flex-col gap-2.5">
      <div className={`${cardSurfaceClass} px-3 py-2.5`}>
        <p className={`${cardTextClass} text-foreground font-semibold`}>
          Capacity check
        </p>
        <p className={`${cardMetaTextClass} text-text-muted mt-1`}>
          Owners, estimates, and timing reviewed together
        </p>
      </div>
      <div className={`${cardSurfaceClass} flex-1 p-3`}>
        <div className="space-y-2.5">
          {[
            ["Operations", "68%", "68%"],
            ["Product", "42%", "42%"],
          ].map(([team, load, width]) => (
            <div key={team}>
              <div className="mb-2 flex items-center justify-between">
                <span
                  className={`${cardMetaTextClass} text-foreground font-medium`}
                >
                  {team}
                </span>
                <span className={`${cardMetaTextClass} text-text-muted`}>
                  {load}
                </span>
              </div>
              <div className="bg-surface-muted h-2 overflow-hidden rounded-full dark:bg-white/10">
                <div
                  className="bg-foreground h-full rounded-full"
                  style={{ width }}
                />
              </div>
            </div>
          ))}
        </div>
        <p
          className={`${cardMetaTextClass} bg-accent text-text-secondary mt-4 rounded-lg px-3 py-2 font-medium`}
        >
          AI recommends Product for the next operations request.
        </p>
      </div>
    </div>
  );
}

function ApprovalControlCard() {
  return (
    <div className="flex h-full flex-col gap-2.5">
      <div className={`${cardSurfaceClass} px-3 py-2.5`}>
        <p className={`${cardTextClass} text-foreground font-semibold`}>
          Review before apply
        </p>
        <p className={`${cardMetaTextClass} text-text-muted mt-1`}>
          Every AI change stays editable before it moves work.
        </p>
      </div>
      <div className={`${cardSurfaceClass} flex-1 p-3`}>
        {[
          ["Owner", "Assign to Maya Chen"],
          ["Estimate", "Set initial estimate to 3 days"],
          ["Start", "Schedule for Thursday morning"],
        ].map(([label, value]) => (
          <div
            className="flex items-center justify-between gap-4 rounded-lg bg-black/4 px-3 py-2 dark:bg-white/7 [&+&]:mt-2"
            key={label}
          >
            <span
              className={`${cardMetaTextClass} text-text-muted font-medium`}
            >
              {label}
            </span>
            <span
              className={`${cardMetaTextClass} text-foreground text-right font-semibold`}
            >
              {value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function DataTable({ rows }: { rows: Array<[string, string]> }) {
  return (
    <div className="border-border my-10 overflow-hidden border-y">
      <table className="w-full text-left text-[0.95rem]">
        <thead>
          <tr className="border-border text-foreground border-b">
            <th className="w-[34%] py-4 pr-6 font-semibold">Input</th>
            <th className="py-4 font-semibold">What FortyOne prepares</th>
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

export default function OperationsUseCasePage() {
  return (
    <>
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
                    FortyOne for Operations
                  </h1>
                </div>
              </div>
              <time
                className="text-text-muted mt-5 block text-sm font-medium"
                dateTime="2026-06-24"
              >
                June 24, 2026
              </time>
            </header>

            <div className="text-text-muted mt-18 text-[1.05rem] leading-8">
              <p className="mb-6">
                Operations work usually starts with a scattered request: a Slack
                thread, a dashboard screenshot, a calendar constraint, and a few
                unanswered questions about who should own the work.
              </p>
              <p className="mb-6">
                FortyOne turns that context into a project plan. It can create
                the task, suggest the right owner, fill in an estimate, plan
                around real capacity, and stop for review before important AI
                actions are applied.
              </p>

              <UseCaseCardGrid>
                <UseCaseFeatureCard>
                  <OperationsPacketCard />
                </UseCaseFeatureCard>
                <UseCaseFeatureCard>
                  <TaskIntakeCard />
                </UseCaseFeatureCard>
              </UseCaseCardGrid>

              <section id="weekly-ops-packet" className="scroll-mt-28">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Weekly ops packet without the manual chase
                </h2>
                <p className="mb-6">
                  Every operations meeting needs the same basic answer: what
                  moved, what is blocked, what needs a decision, and who has
                  enough room to take the next piece of work.
                </p>
                <p className="mb-6">
                  FortyOne keeps that packet connected to the actual project
                  plan, so the review is not rebuilt from scratch in slides,
                  spreadsheets, and status messages.
                </p>
                <DataTable rows={packetRows} />
                <PromptCard title="Weekly operations packet">
                  Build this week&apos;s operations packet from active projects,
                  overdue tasks, blocked work, and team capacity. Include source
                  links and stop before changing any assignments.
                </PromptCard>
              </section>

              <section id="task-intake" className="scroll-mt-28">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Turn requests into assigned work
                </h2>
                <p className="mb-6">
                  Requests often arrive before they are ready to assign.
                  FortyOne can turn rough notes into a structured task, connect
                  it to a goal, and recommend who should own it based on
                  workload and context.
                </p>
                <DataTable rows={intakeRows} />
                <PromptCard title="Task intake">
                  Create a task from this Slack request, attach the relevant
                  goal, suggest the best owner, estimate the work, and choose a
                  schedule window that does not overload the team.
                </PromptCard>
              </section>

              <section id="capacity" className="scroll-mt-28">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Plan around real team capacity
                </h2>
                <p className="mb-6">
                  The best assignment is not just the person who can do the
                  work. It is the person who can do it at the right time without
                  creating a hidden bottleneck somewhere else.
                </p>
                <UseCaseCardGrid>
                  <UseCaseFeatureCard>
                    <CapacityPlanCard />
                  </UseCaseFeatureCard>
                  <UseCaseFeatureCard>
                    <ApprovalControlCard />
                  </UseCaseFeatureCard>
                </UseCaseCardGrid>
              </section>

              <section id="approval-control" className="scroll-mt-28">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Keep AI actions under review
                </h2>
                <p className="mb-6">
                  AI should help prepare the plan, not quietly change how teams
                  work. FortyOne can stage owners, estimates, schedule changes,
                  and next tasks for review before they are applied.
                </p>
                <PromptCard title="Review before apply">
                  Review the AI plan for this workstream. Show the proposed
                  owner, estimate, start time, and task changes. Let me edit or
                  approve the changes before anything moves.
                </PromptCard>
              </section>

              <section id="questions" className="scroll-mt-28">
                <h2 className="text-foreground mt-14 mb-5 text-2xl font-semibold tracking-normal md:text-[1.7rem]">
                  Questions operations teams ask
                </h2>
                <div className="divide-border border-border divide-y border-y">
                  {[
                    [
                      "Can FortyOne create tasks from Slack?",
                      "Yes. Slack can become an intake path for project work, with AI helping turn conversations into structured tasks.",
                    ],
                    [
                      "Can AI assign work automatically?",
                      "AI can suggest owners, estimates, and timing. Teams can keep review controls in place before important changes are applied.",
                    ],
                    [
                      "Does this replace project managers?",
                      "No. It removes coordination drag so managers can spend more time deciding priorities and less time rebuilding status.",
                    ],
                  ].map(([question, answer]) => (
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
