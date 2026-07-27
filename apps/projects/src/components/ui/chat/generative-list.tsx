"use client";

import type { ReactNode } from "react";
import { Children, useState } from "react";
import Link from "next/link";

export const GENERATIVE_LIST_PREVIEW_LIMIT = 5;

export const GenerativeList = ({
  children,
  emptyMessage,
  title,
}: {
  children: ReactNode;
  emptyMessage: string;
  title: string;
}) => {
  const [showAll, setShowAll] = useState(false);
  const items = Children.toArray(children);
  const visibleItems = showAll
    ? items
    : items.slice(0, GENERATIVE_LIST_PREVIEW_LIMIT);
  const remainingItemCount = items.length - GENERATIVE_LIST_PREVIEW_LIMIT;
  const hasAdditionalItems = remainingItemCount > 0;

  if (items.length === 0) {
    return <p className="text-text-muted text-base">{emptyMessage}</p>;
  }

  return (
    <section
      aria-label={title}
      className="mt-3 grid w-full max-w-full min-w-0 gap-2.5 overflow-hidden"
    >
      <span className="text-text-muted text-base font-medium">{title}</span>
      <div className="border-border/70 grid w-full max-w-full min-w-0 overflow-hidden border-t dark:border-white/[0.09]">
        {visibleItems}
        {hasAdditionalItems ? (
          <button
            className="border-border/70 text-foreground focus-visible:ring-foreground/50 min-h-12 w-full border-b px-px py-3 text-left text-base font-medium transition-opacity outline-none hover:opacity-70 focus-visible:ring-2 focus-visible:ring-offset-2 dark:border-white/[0.09]"
            onClick={() => {
              setShowAll((current) => !current);
            }}
            type="button"
          >
            {showAll ? "Show less" : `View ${remainingItemCount} more`}
          </button>
        ) : null}
      </div>
    </section>
  );
};

const itemClassName =
  "border-border/70 text-foreground focus-visible:ring-foreground/50 flex min-h-[52px] w-full max-w-full min-w-0 items-center gap-3 overflow-hidden border-b px-px py-3 text-base leading-6 no-underline transition-[opacity,transform] duration-150 outline-none dark:border-white/[0.09]";

export const GenerativeListItem = ({
  ariaLabel,
  href,
  leading,
  title,
  trailing,
}: {
  ariaLabel?: string;
  href?: string;
  leading: ReactNode;
  title: string;
  trailing?: ReactNode;
}) => {
  const content = (
    <>
      <span className="flex size-5 shrink-0 items-center justify-center">
        {leading}
      </span>
      <span className="min-w-0 flex-1 truncate">{title}</span>
      {trailing ? (
        <span className="text-text-muted flex min-w-0 shrink-0 items-center justify-end text-base">
          {trailing}
        </span>
      ) : null}
    </>
  );

  if (!href) {
    return <div className={itemClassName}>{content}</div>;
  }

  const interactiveClassName = `${itemClassName} hover:opacity-70 focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.99]`;
  const isExternal = href.startsWith("http://") || href.startsWith("https://");

  if (isExternal) {
    return (
      <a
        aria-label={ariaLabel ?? `Open ${title}`}
        className={interactiveClassName}
        href={href}
        rel="noreferrer"
        target="_blank"
      >
        {content}
      </a>
    );
  }

  return (
    <Link
      aria-label={ariaLabel ?? `Open ${title}`}
      className={interactiveClassName}
      href={href}
    >
      {content}
    </Link>
  );
};
