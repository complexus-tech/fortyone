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

export function FeedbackPreview() {
  return (
    <GlassFrame
      footer={<span className={styles.goal}>Goal · Activation</span>}
      title="Customer feedback"
    >
      <div className={styles.row}>
        <span className={cn("rounded-lg", styles.icon, styles.coral)}>
          <RequestsIcon className="size-4 text-current" />
        </span>
        <div>
          <p className={styles.strong}>Make onboarding easier</p>
          <p className={styles.meta}>12 votes</p>
        </div>
      </div>
      <div className={styles.work}>
        <div className={styles.rowBetween}>
          <span className={styles.meta}>Planned task</span>
          <span className={styles.meta}>PRD-142</span>
        </div>
        <p className={styles.task}>Redesign onboarding flow</p>
        <span className={cn("rounded-lg", styles.status)}>Planned</span>
      </div>
      <p className={cn(styles.meta, styles.note)}>Original request attached</p>
    </GlassFrame>
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

export function ContextPreview() {
  return (
    <GlassFrame
      footer="Manage tools"
      title={
        <>
          <span>Add context from</span>
          <span>@</span>
        </>
      }
    >
      <div className={styles.integration}>
        <span className={cn("rounded-lg", styles.icon)}>
          <GitHubIcon className="size-5" />
        </span>
        <p className={styles.strong}>GitHub</p>
      </div>
      <div className={styles.integration}>
        <span className={cn("rounded-lg", styles.icon)}>
          <Image
            alt=""
            className="size-5 object-contain"
            height={20}
            src="/integrations/google-calendar-2026.svg"
            width={20}
          />
        </span>
        <p className={styles.strong}>Google Calendar</p>
      </div>
      <div className={styles.integration}>
        <span className={cn("rounded-lg", styles.icon, styles.coral)}>
          <SlackIcon className="size-5" />
        </span>
        <p className={styles.strong}>Slack</p>
        <span className={cn(styles.meta, styles.connect)}>Connect</span>
      </div>
    </GlassFrame>
  );
}
