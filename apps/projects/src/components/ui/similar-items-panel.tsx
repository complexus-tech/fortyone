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
      className="bg-surface-elevated border-border-strong shadow-shadow dark:bg-surface-elevated/80 absolute top-[calc(100%+0.75rem)] left-0 w-full overflow-hidden rounded-2xl border-[0.5px] shadow-2xl backdrop-blur-xl"
    >
      <div className="border-border-strong bg-surface-muted/70 text-text-muted flex min-h-16 items-center border-b-[0.5px] px-6 py-4 text-sm font-semibold tracking-[0.08em] uppercase dark:bg-white/[0.035]">
        {heading}
      </div>
      <div className="divide-border-strong divide-y-[0.5px]">{children}</div>
    </section>
  );
};
