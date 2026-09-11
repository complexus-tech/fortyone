import Image from "next/image";
import {
  AiIcon,
  CheckIcon,
  ClockIcon,
  DocsIcon,
  GoalIcon,
  LinkIcon,
  LockIcon,
  TeamIcon,
} from "icons";
import { cn } from "lib";
import { PricingFeedbackIcon } from "@/components/ui/pricing-icons";
import { IntegrationBrand } from "@/modules/integrations/integration-brand";
import avatarStyles from "@/modules/home/capability-previews.module.css";
import josephAvatar from "../../../public/joseph.webp";
import tasksPhoto from "../../../public/images/workflows/tasks.webp";
import goalsPhoto from "../../../public/images/workflows/goals.webp";
import roadmapsPhoto from "../../../public/images/workflows/roadmaps.webp";
import feedbackPhoto from "../../../public/images/workflows/customer-feedback.webp";
import supportPhoto from "../../../public/images/workflows/customer-support.webp";
import constructionPhoto from "../../../public/images/workflows/construction.webp";
import planningPhoto from "../../../public/images/workflows/ai-planning.webp";
import type {
  CapacityPreview,
  HandoffConfig,
  HandoffIcon,
  HandoffMark,
  HandoffPreview as HandoffPreviewData,
  IntakePreview,
  ReviewPreview,
} from "./handoff-types";
import styles from "./handoffs.module.css";

const BACKDROPS = {
  tasks: tasksPhoto,
  goals: goalsPhoto,
  roadmaps: roadmapsPhoto,
  "customer-feedback": feedbackPhoto,
  "customer-support": supportPhoto,
  construction: constructionPhoto,
  "ai-planning": planningPhoto,
};

const HANDOFF_ICONS = {
  document: DocsIcon,
  clock: ClockIcon,
  link: LinkIcon,
  team: TeamIcon,
  goal: GoalIcon,
  lock: LockIcon,
  feedback: PricingFeedbackIcon,
  check: CheckIcon,
} satisfies Record<HandoffIcon, typeof DocsIcon>;

const TONES = {
  sky: styles.skyIcon,
  sage: styles.sageIcon,
  amber: styles.amberIcon,
};

function isHandoffIcon(mark: HandoffMark): mark is HandoffIcon {
  return mark in HANDOFF_ICONS;
}

function PreviewMark({
  mark,
  size = 16,
  className,
}: {
  mark: HandoffMark;
  size?: number;
  className?: string;
}) {
  if (mark === "joseph") {
    return (
      <span className={cn("rounded-full", styles.avatar)}>
        <Image
          alt=""
          className={avatarStyles.avatarImage}
          height={28}
          src={josephAvatar}
          width={28}
        />
      </span>
    );
  }
  if (isHandoffIcon(mark)) {
    const Icon = HANDOFF_ICONS[mark];
    return (
      <Icon
        className={cn("shrink-0 text-current", className)}
        strokeWidth={1.75}
        style={{ width: size, height: size }}
      />
    );
  }
  return <IntegrationBrand name={mark} size={size} />;
}

function ReviewSheet({ preview }: { preview: ReviewPreview }) {
  return (
    <>
      <div className={cn("rounded-xl", styles.sheet, styles.brief)}>
        <div className={styles.toolbar}>
          <span className={styles.inline}>
            <PreviewMark className={styles.skyIcon} mark="document" />
            {preview.toolbar}
          </span>
          <span className={styles.meta}>{preview.status}</span>
        </div>
        <div className={styles.documentBody}>
          <p className={styles.meta}>{preview.eyebrow}</p>
          <h4 className="mt-3 text-lg font-semibold">{preview.title}</h4>
          <p className={cn(styles.meta, "mt-2")}>{preview.description}</p>
          <div className={styles.briefRows}>
            {preview.rows.map((row) => (
              <div key={row.label}>
                <PreviewMark className={TONES[row.tone]} mark={row.icon} />
                <span>{row.label}</span>
                <span
                  className={
                    row.tone === "amber" ? styles.riskText : styles.meta
                  }
                >
                  {row.value}
                </span>
              </div>
            ))}
          </div>
          <div className={styles.decisionNote}>
            <p className={styles.meta}>NEXT DECISION</p>
            <p className="mt-2 font-medium">{preview.decision.title}</p>
            <p className={cn(styles.meta, "mt-2")}>
              {preview.decision.description}
            </p>
          </div>
        </div>
        <div className={styles.sheetFooter}>
          <span className={styles.inline}>
            {preview.sources.map((source) => (
              <PreviewMark key={source} mark={source} />
            ))}
          </span>
          <span className={styles.meta}>Source links included</span>
        </div>
      </div>
      <div className={cn("rounded-lg", styles.annotation)}>
        <CheckIcon
          className={cn("size-4 shrink-0", styles.sageIcon)}
          strokeWidth={1.75}
        />
        <span>
          {preview.note.title}{" "}
          <span className={styles.meta}>{preview.note.detail}</span>
        </span>
      </div>
    </>
  );
}

function IntakeCards({ preview }: { preview: IntakePreview }) {
  return (
    <>
      <div className={cn("rounded-xl", styles.sheet, styles.message)}>
        <div className={styles.inline}>
          <PreviewMark mark={preview.source.icon} size={20} />
          <span className="font-medium">{preview.source.label}</span>
          <span className={cn(styles.meta, "ml-auto")}>
            {preview.source.badge}
          </span>
        </div>
        <div className={cn(styles.inline, "mt-5 items-start")}>
          <PreviewMark
            className={styles.skyIcon}
            mark={preview.source.authorIcon}
            size={20}
          />
          <div>
            <p className="font-medium">{preview.source.author}</p>
            <p className="mt-2">{preview.source.message}</p>
          </div>
        </div>
      </div>
      <div className={styles.connector}>
        <svg
          className={styles.handoffArrow}
          fill="none"
          preserveAspectRatio="none"
          viewBox="0 0 100 80"
        >
          <path
            d="M0 4C-13 44 8 62 100 63M90 54L100 63L90 69"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="1.75"
          />
        </svg>
        <span className={cn("rounded-md", styles.inline)}>
          Context travels with the work
        </span>
      </div>
      <div className={cn("rounded-xl", styles.sheet, styles.task)}>
        <div className={styles.toolbar}>
          <span className={styles.meta}>{preview.task.identifier}</span>
          <span className={styles.inline}>
            <span className={styles.statusDot} />
            Needs review
          </span>
        </div>
        <div className={styles.taskBody}>
          <h4 className="text-lg font-semibold">{preview.task.title}</h4>
          <p className={cn(styles.meta, "mt-2")}>{preview.task.description}</p>
          <dl className={styles.taskFields}>
            {preview.task.fields.map((field) => (
              <div key={field.label}>
                <dt>{field.label}</dt>
                <dd className={styles.inline}>
                  <PreviewMark mark={field.icon} />
                  <span>{field.value}</span>
                </dd>
              </div>
            ))}
          </dl>
        </div>
        <div className={styles.sheetFooter}>
          <span className={styles.inline}>
            <PreviewMark className={styles.skyIcon} mark="link" />
            {preview.sourceCaption}
          </span>
          <PreviewMark mark={preview.source.icon} />
        </div>
      </div>
    </>
  );
}

function CapacityChart({ preview }: { preview: CapacityPreview }) {
  return (
    <>
      <div className={cn("rounded-xl", styles.sheet, styles.capacity)}>
        <div className={styles.toolbar}>
          <span className={cn("font-medium", styles.inline)}>
            <PreviewMark className={styles.skyIcon} mark="team" />
            {preview.toolbar}
          </span>
          <span className={styles.meta}>This week</span>
        </div>
        <div className={styles.capacityBody}>
          <div className={styles.chartScale}>
            <span>Allocated time</span>
            <span>100%</span>
          </div>
          {preview.allocations.map((allocation, index) => (
            <div
              className={cn(
                styles.allocation,
                index === 1 && styles.recommended,
              )}
              key={allocation.team}
            >
              <div>
                <span className={styles.teamLabel}>{allocation.team}</span>
                <span>{allocation.percent}%</span>
              </div>
              <div className={styles.track}>
                <span style={{ width: `${allocation.percent}%` }} />
              </div>
              {index === 1 ? (
                <p className={cn(styles.meta, "mt-3")}>
                  {preview.recommendation}
                </p>
              ) : null}
            </div>
          ))}
          <div className={styles.calendarHeading}>
            <span className="font-medium">A realistic start window</span>
            <span className={styles.inline}>
              <IntegrationBrand name="Google Calendar" size={16} />
              <IntegrationBrand name="Outlook Calendar" size={16} />
            </span>
          </div>
          <div className={styles.week}>
            {["Mon", "Tue", "Wed", "Thu", "Fri"].map((day) => (
              <div key={day}>
                <span className={styles.meta}>{day}</span>
                <span
                  className={cn(
                    "rounded-md",
                    styles.dayBlock,
                    day === preview.startDay && styles.selectedDay,
                  )}
                >
                  {day === preview.startDay ? "Start" : ""}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className={cn("rounded-xl", styles.sheet, styles.proposal)}>
        <span className={styles.proposalMark}>
          <AiIcon className="size-5 text-current" strokeWidth={1.75} />
        </span>
        <div>
          <p className="font-medium">Propose it. Review it. Then apply.</p>
          <p className={cn(styles.meta, "mt-1")}>{preview.proposal}</p>
        </div>
      </div>
    </>
  );
}

export function HandoffPreview({
  preview,
  image,
}: {
  preview: HandoffPreviewData;
  image: HandoffConfig["image"];
}) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        styles.scene,
        preview.kind === "intake" && styles.intakeScene,
      )}
    >
      <div className={cn("rounded-2xl", styles.backdrop)}>
        <Image
          alt=""
          className={styles.backdropImage}
          fill
          placeholder="blur"
          sizes="(max-width: 767px) 100vw, (max-width: 1023px) 720px, 680px"
          src={BACKDROPS[image]}
        />
      </div>
      {preview.kind === "review" && <ReviewSheet preview={preview} />}
      {preview.kind === "intake" && <IntakeCards preview={preview} />}
      {preview.kind === "capacity" && <CapacityChart preview={preview} />}
    </div>
  );
}
