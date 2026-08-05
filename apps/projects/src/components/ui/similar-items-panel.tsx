import type { ReactNode } from "react";
import { ArrowRight2Icon } from "icons";
import { cn } from "lib";

export type SimilarItemPanelItem = {
  id: string;
  title: string;
  meta?: ReactNode;
  label?: string;
  isBlocking?: boolean;
};

export const SimilarItemsPanel = ({
  heading,
  items,
  onSelect,
  statusMessage,
  statusTone = "muted",
}: {
  heading: string;
  items: SimilarItemPanelItem[];
  onSelect: (item: SimilarItemPanelItem) => void;
  statusMessage?: string;
  statusTone?: "muted" | "error";
}) => {
  if (items.length === 0 && !statusMessage) return null;

  return (
    <section
      aria-label={heading}
      className="bg-surface-elevated/95 border-border/70 shadow-shadow absolute top-[calc(100%+0.75rem)] left-0 w-full overflow-hidden rounded-2xl border-[0.5px] shadow-2xl backdrop-blur-xl"
    >
      <div className="border-border/70 text-text-muted border-b-[0.5px] px-6 py-3 text-xs font-semibold tracking-[0.08em] uppercase">
        {heading}
      </div>
      <div className="divide-border/60 divide-y-[0.5px]">
        {statusMessage ? (
          <p
            className={cn(
              "px-6 py-4 text-sm",
              statusTone === "error" ? "text-danger" : "text-text-muted",
            )}
          >
            {statusMessage}
          </p>
        ) : null}
        {!statusMessage
          ? items.map((item) => (
              <button
                className="hover:bg-state-hover/40 focus-visible:bg-state-hover/40 group flex w-full items-center gap-4 px-6 py-4 text-left transition-colors outline-none"
                key={item.id}
                onClick={() => {
                  onSelect(item);
                }}
                type="button"
              >
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-[1rem] font-semibold">
                    {item.title}
                  </span>
                  {item.meta ? (
                    <span className="text-text-muted mt-1.5 flex items-center gap-3 text-sm">
                      {item.meta}
                    </span>
                  ) : null}
                </span>
                {item.label ? (
                  <span
                    className={cn(
                      "shrink-0 rounded-lg border px-2 py-1 text-xs font-medium",
                      item.isBlocking
                        ? "border-warning/30 bg-warning/10 text-warning"
                        : "border-border/70 text-text-muted",
                    )}
                  >
                    {item.label}
                  </span>
                ) : null}
                <ArrowRight2Icon className="text-text-muted h-4.5 shrink-0 opacity-50 transition-transform group-hover:translate-x-0.5 group-hover:opacity-100" />
              </button>
            ))
          : null}
      </div>
    </section>
  );
};
