import type { ComponentPropsWithoutRef } from "react";
import { cn } from "lib";

export type HandwrittenAccentTone = "danger" | "primary" | "success";

type HandwrittenAccentProps = ComponentPropsWithoutRef<"span"> & {
  tone: HandwrittenAccentTone;
};

const TONE_CLASS_NAMES: Record<HandwrittenAccentTone, string> = {
  danger: "text-danger",
  primary: "text-primary",
  success: "text-success",
};

const INK_DROP_CLASS_NAMES: Record<HandwrittenAccentTone, readonly string[]> = {
  primary: [
    "-top-[0.11em] left-[20%] size-[0.045em] opacity-55",
    "-top-[0.04em] -right-[0.1em] h-[0.085em] w-[0.05em] rotate-[18deg] opacity-65",
    "bottom-[0.02em] -left-[0.07em] size-[0.045em] opacity-50",
    "-top-[0.1em] right-[17%] size-[0.03em] opacity-45",
    "-bottom-[0.01em] right-[3%] size-[0.025em] opacity-40",
  ],
  danger: [
    "-top-[0.08em] left-[7%] h-[0.05em] w-[0.075em] -rotate-12 opacity-50",
    "top-[0.1em] -right-[0.1em] size-[0.065em] opacity-65",
    "-bottom-[0.01em] left-[52%] size-[0.035em] opacity-45",
    "bottom-[0.03em] -left-[0.08em] h-[0.08em] w-[0.045em] rotate-12 opacity-65",
    "-top-[0.12em] right-[22%] size-[0.028em] opacity-45",
  ],
  success: [
    "-top-[0.1em] left-[30%] size-[0.055em] opacity-55",
    "top-[0.12em] -right-[0.11em] h-[0.045em] w-[0.075em] rotate-[20deg] opacity-55",
    "bottom-[0.02em] -left-[0.08em] size-[0.055em] opacity-65",
    "-top-[0.04em] -left-[0.08em] size-[0.03em] opacity-45",
    "-bottom-[0.02em] right-[18%] size-[0.03em] opacity-45",
  ],
};

const InkDrops = ({ tone }: { tone: HandwrittenAccentTone }) => {
  return (
    <span
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 mix-blend-multiply filter-[brightness(0.88)] dark:mix-blend-normal dark:filter-[brightness(0.68)]"
    >
      {INK_DROP_CLASS_NAMES[tone].map((className) => (
        <span
          className={cn(
            "absolute rounded-[55%_45%_60%_40%] bg-current",
            className,
          )}
          key={className}
        />
      ))}
    </span>
  );
};

const HandwrittenUnderline = () => {
  return (
    <svg
      aria-hidden="true"
      className="pointer-events-none absolute -bottom-[0.02em] left-0 h-[0.16em] w-full overflow-visible"
      fill="none"
      preserveAspectRatio="none"
      viewBox="0 0 100 12"
    >
      <path
        d="M2 7.8C12 5.2 19 8.6 30 6.5C42 4.2 51 8.4 62 6.1C74 3.8 86 7.9 98 5.2"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="2.4"
        vectorEffect="non-scaling-stroke"
      />
      <path
        d="M5 10.1C22 8.4 38 10.4 55 8.3C70 6.5 84 9.4 96 7.6"
        opacity="0.38"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="1.2"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
};

export const HandwrittenAccent = ({
  children,
  className,
  tone,
  ...props
}: HandwrittenAccentProps) => {
  return (
    <span
      className={cn(
        "font-handwritten relative inline-block text-[1.08em] leading-[0.9] tracking-[-0.045em]",
        TONE_CLASS_NAMES[tone],
        className,
      )}
      {...props}
    >
      {children}
      <InkDrops tone={tone} />
    </span>
  );
};

export const UnderlinedHandwrittenAccent = ({
  children,
  ...props
}: HandwrittenAccentProps) => {
  return (
    <HandwrittenAccent {...props}>
      {children}
      <HandwrittenUnderline />
    </HandwrittenAccent>
  );
};
