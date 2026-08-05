import { Children, type ReactNode } from "react";

export const SimilarItemsPanel = ({
  children,
  heading,
}: {
  children: ReactNode;
  heading: string;
}) => {
  if (Children.count(children) === 0) return null;

  return (
    <section
      aria-label={heading}
      className="bg-surface-elevated border-border-strong shadow-shadow absolute top-[calc(100%+0.75rem)] left-0 w-full overflow-hidden rounded-2xl border-[0.5px] shadow-2xl"
    >
      <div className="border-border-strong bg-surface-muted/25 text-text-muted flex min-h-12 items-center border-b-[0.5px] px-6 py-3 text-sm font-semibold tracking-[0.08em] uppercase dark:bg-white/[0.01]">
        {heading}
      </div>
      <div className="divide-border-strong divide-y-[0.5px]">{children}</div>
    </section>
  );
};
