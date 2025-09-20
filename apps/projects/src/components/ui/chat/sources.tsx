"use client";

import { ArrowDown2Icon, BookIcon, InternetIcon } from "icons";
import { cn } from "lib";
import type { ComponentProps } from "react";
import { Collapsible } from "ui";

type SourcesProps = ComponentProps<"div">;

export const Sources = ({ className, ...props }: SourcesProps) => (
  <Collapsible className={cn("not-prose mb-4 mt-2", className)} {...props} />
);

type SourcesTriggerProps = ComponentProps<typeof Collapsible.Trigger> & {
  count: number;
};

export const Trigger = ({ count, children, ...props }: SourcesTriggerProps) => (
  <Collapsible.Trigger className="flex items-center gap-1" {...props}>
    {children ?? (
      <>
        <p className="flex items-center gap-1.5 text-[0.95rem] font-medium">
          <InternetIcon className="h-4 text-dark dark:text-white" />
          Used {count} sources
        </p>
        <ArrowDown2Icon className="h-4 text-dark dark:text-white" />
      </>
    )}
  </Collapsible.Trigger>
);

export type SourcesContentProps = ComponentProps<typeof Collapsible.Content>;

const Content = ({ className, ...props }: SourcesContentProps) => (
  <Collapsible.Content
    className={cn(
      "mt-3 flex w-fit flex-col gap-2 pl-4 outline-none",
      className,
    )}
    {...props}
  />
);

type SourceProps = ComponentProps<"a">;

const Source = ({ href, title, children, ...props }: SourceProps) => (
  <a
    className="flex items-center gap-1.5 underline underline-offset-2"
    href={href}
    rel="noreferrer"
    target="_blank"
    {...props}
  >
    {children ?? (
      <>
        <BookIcon className="relative top-[0.5px] h-4 text-dark dark:text-white" />
        <span className="block font-medium">{title}</span>
      </>
    )}
  </a>
);

Sources.Content = Content;
Sources.Trigger = Trigger;
Sources.Source = Source;
