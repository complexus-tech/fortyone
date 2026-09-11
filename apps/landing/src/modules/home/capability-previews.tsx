import type { ReactNode } from "react";
import { AiIcon, RequestsIcon, GitHubIcon } from "icons";
import { cn } from "lib";
import Image from "next/image";
import josephAvatar from "../../../public/joseph.webp";
import { SlackIcon } from "./how-it-works";
import styles from "./capability-previews.module.css";

function GlassFrame({
  children,
  footer,
  title,
}: {
  children: ReactNode;
  footer: ReactNode;
  title: ReactNode;
}) {
  return (
    <div className={cn("rounded-2xl", styles.frame)}>
      <div className={styles.header}>{title}</div>
      <div className={cn("rounded-xl", styles.surface)}>{children}</div>
      <div className={styles.footer}>{footer}</div>
    </div>
  );
}

export type EvidencePreviewContent = {
  title: ReactNode;
  icon: ReactNode;
  request: string;
  context: string;
  workLabel: string;
  workReference?: string;
  workTitle: string;
  status?: string;
  note: string;
  footer: ReactNode;
};

// The same preview is used on the homepage and detail pages; only content varies.
export function EvidencePreview({
  title,
  icon,
  request,
  context,
  workLabel,
  workReference,
  workTitle,
  status,
  note,
  footer,
}: EvidencePreviewContent) {
  return (
    <GlassFrame footer={footer} title={title}>
      <div className={styles.row}>
        <span className={cn("rounded-lg", styles.icon, styles.coral)}>
          {icon}
        </span>
        <div>
          <p className={styles.strong}>{request}</p>
          <p className={styles.meta}>{context}</p>
        </div>
      </div>
      <div className={styles.work}>
        <div className={styles.rowBetween}>
          <span className={styles.meta}>{workLabel}</span>
          <span className={styles.meta}>{workReference}</span>
        </div>
        <p className={styles.task}>{workTitle}</p>
        {status ? (
          <span className={cn("rounded-lg", styles.status)}>{status}</span>
        ) : null}
      </div>
      <p className={cn(styles.meta, styles.note)}>{note}</p>
    </GlassFrame>
  );
}

export function FeedbackPreview() {
  return (
    <EvidencePreview
      context="12 votes"
      footer={<span className={styles.goal}>Goal · Activation</span>}
      icon={<RequestsIcon className="size-4 text-current" />}
      note="Original request attached"
      request="Make onboarding easier"
      status="Planned"
      title="Customer feedback"
      workLabel="Planned task"
      workReference="PRD-142"
      workTitle="Redesign onboarding flow"
    />
  );
}

export function PlanningPreview() {
  return (
    <GlassFrame
      footer={
        <>
          <span className="flex shrink-0 items-center gap-1">
            <Image
              alt="Google Calendar"
              className="size-4 object-contain"
              height={16}
              src="/integrations/google-calendar-2026.svg"
              width={16}
            />
            <Image
              alt="Outlook Calendar"
              className="size-4 object-contain"
              height={16}
              src="/integrations/outlook-2025.svg"
              width={16}
            />
          </span>
          <span>Calendar and workload checked</span>
        </>
      }
      title={
        <>
          <AiIcon className="size-4 text-current" />
          Maya&apos;s work plan
        </>
      }
    >
      <p className={cn(styles.meta, styles.note)}>
        Task: redesign onboarding flow
      </p>
      <div className={styles.owner}>
        <span className={cn("rounded-full", styles.avatar)}>
          <Image
            alt="Joseph"
            className={styles.avatarImage}
            height={36}
            sizes="36px"
            src={josephAvatar}
            width={36}
          />
        </span>
        <div>
          <p className={styles.meta}>Suggested owner</p>
          <p className={styles.strong}>Joseph</p>
        </div>
        <span className={cn("rounded-lg", styles.status)}>Review</span>
      </div>
      <div className={styles.schedule}>
        <div>
          <p className={styles.meta}>Planned effort</p>
          <p className={styles.strong}>4 hours</p>
        </div>
        <div>
          <p className={styles.meta}>First work block</p>
          <p className={styles.strong}>Tue 10:30</p>
        </div>
      </div>
    </GlassFrame>
  );
}

type ContextPreviewItem = {
  id?: string;
  label: string;
  icon: ReactNode;
  detail?: string;
  trailing?: string;
  coral?: boolean;
};

export function ContextListPreview({
  title,
  footer,
  items,
}: {
  title: ReactNode;
  footer: ReactNode;
  items: readonly ContextPreviewItem[];
}) {
  return (
    <GlassFrame footer={footer} title={title}>
      {items.map(({ id, label, icon, detail, trailing, coral }) => (
        <div className={styles.integration} key={id ?? label}>
          <span
            className={cn("rounded-lg", styles.icon, coral && styles.coral)}
          >
            {icon}
          </span>
          <div>
            <p className={styles.strong}>{label}</p>
            {detail ? <p className={styles.meta}>{detail}</p> : null}
          </div>
          {trailing ? (
            <span className={cn(styles.meta, styles.connect)}>{trailing}</span>
          ) : null}
        </div>
      ))}
    </GlassFrame>
  );
}

export function ContextPreview() {
  return (
    <ContextListPreview
      footer="Manage tools"
      items={[
        { label: "GitHub", icon: <GitHubIcon className="size-5" /> },
        {
          label: "Google Calendar",
          icon: (
            <Image
              alt=""
              className="size-5 object-contain"
              height={20}
              src="/integrations/google-calendar-2026.svg"
              width={20}
            />
          ),
        },
        {
          label: "Slack",
          icon: <SlackIcon className="size-5" />,
          trailing: "Connect",
          coral: true,
        },
      ]}
      title={
        <>
          <span>Add context from</span>
          <span>@</span>
        </>
      }
    />
  );
}
