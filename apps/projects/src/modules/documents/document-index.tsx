"use client";

import { useEffect, useState } from "react";
import { Button, Popover, Text } from "ui";
import { ListIcon } from "icons";
import { cn } from "lib";
import styles from "./document-index.module.css";

type Heading = {
  element: HTMLElement;
  id: number;
  level: number;
  text: string;
};

function collectDocumentHeadings(container: HTMLElement, selector: string) {
  const content = container.querySelector(selector);
  if (!content) return [];
  return Array.from(
    content.querySelectorAll<HTMLElement>("h1, h2, h3, h4, h5, h6"),
  ).filter((element) => element.textContent.trim());
}

export function DocumentIndex({
  scrollContainer,
  contentSelector,
  readingOffset = 32,
}: {
  scrollContainer: HTMLElement | null;
  contentSelector: string;
  readingOffset?: number;
}) {
  const [headings, setHeadings] = useState<Heading[]>([]);
  const [active, setActive] = useState<HTMLElement | null>(null);
  const [open, setOpen] = useState(false);
  const [compactPosition, setCompactPosition] = useState({
    bottom: 24,
    left: 24,
  });

  useEffect(() => {
    if (!scrollContainer) return;
    const measure = () => {
      const bounds = scrollContainer.getBoundingClientRect();
      const left = Number.isFinite(bounds.left) ? bounds.left : 0;
      const bottom = Number.isFinite(bounds.bottom)
        ? bounds.bottom
        : window.innerHeight;
      setCompactPosition({
        bottom: Math.max(16, window.innerHeight - bottom + 24),
        left: Math.max(16, left + 24),
      });
    };
    const resizeObserver = new ResizeObserver(measure);
    resizeObserver.observe(scrollContainer);
    window.addEventListener("resize", measure);
    measure();
    return () => {
      resizeObserver.disconnect();
      window.removeEventListener("resize", measure);
    };
  }, [scrollContainer]);

  useEffect(() => {
    if (!scrollContainer) return;
    let current: Heading[] = [];
    let frame: number | null = null;
    let needsCollect = true;
    let nextId = 0;
    let observedContent: Element | null = null;
    const ids = new WeakMap<HTMLElement, number>();

    const measure = () => {
      frame = null;
      if (needsCollect) {
        needsCollect = false;
        const content = scrollContainer.querySelector(contentSelector);
        if (content !== observedContent) {
          if (observedContent) resizeObserver.unobserve(observedContent);
          if (content) resizeObserver.observe(content);
          observedContent = content;
        }
        const next = collectDocumentHeadings(
          scrollContainer,
          contentSelector,
        ).map((element) => {
          let id = ids.get(element);
          if (id === undefined) {
            id = nextId;
            nextId += 1;
            ids.set(element, id);
          }
          return {
            element,
            id,
            level: Number(element.tagName.slice(1)),
            text: element.textContent.trim(),
          };
        });
        current = next;
        setHeadings((previous) =>
          next.length === previous.length &&
          next.every(
            (heading, index) =>
              heading.element === previous[index].element &&
              heading.text === previous[index].text &&
              heading.level === previous[index].level,
          )
            ? previous
            : next,
        );
      }
      const top =
        scrollContainer.getBoundingClientRect().top + readingOffset + 2;
      let nextActive: HTMLElement | null = current.at(0)?.element ?? null;
      for (const heading of current) {
        if (heading.element.getBoundingClientRect().top > top) break;
        nextActive = heading.element;
      }
      if (
        scrollContainer.scrollTop > 0 &&
        scrollContainer.scrollHeight -
          scrollContainer.clientHeight -
          scrollContainer.scrollTop <=
          2
      ) {
        nextActive = current.at(-1)?.element ?? null;
      }
      setActive((previous) =>
        previous === nextActive ? previous : nextActive,
      );
    };
    const schedule = () => {
      if (frame === null) frame = requestAnimationFrame(measure);
    };
    const observer = new MutationObserver(() => {
      needsCollect = true;
      schedule();
    });
    observer.observe(scrollContainer, {
      childList: true,
      subtree: true,
      characterData: true,
    });
    const resizeObserver = new ResizeObserver(schedule);
    resizeObserver.observe(scrollContainer);
    scrollContainer.addEventListener("scroll", schedule, { passive: true });
    scrollContainer.addEventListener("load", schedule, true);
    schedule();
    return () => {
      observer.disconnect();
      resizeObserver.disconnect();
      scrollContainer.removeEventListener("scroll", schedule);
      scrollContainer.removeEventListener("load", schedule, true);
      if (frame !== null) cancelAnimationFrame(frame);
    };
  }, [scrollContainer, contentSelector, readingOffset]);

  if (headings.length < 2) return null;

  const navigate = (heading: Heading) => {
    if (!scrollContainer) return;
    scrollContainer.scrollTo({
      top: Math.max(
        0,
        scrollContainer.scrollTop +
          heading.element.getBoundingClientRect().top -
          scrollContainer.getBoundingClientRect().top -
          readingOffset,
      ),
      behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches
        ? "instant"
        : "smooth",
    });
    setActive(heading.element);
    setOpen(false);
  };

  const minimumLevel = Math.min(...headings.map((heading) => heading.level));
  const links = (
    <ol className="space-y-1">
      {headings.map((heading) => (
        <li
          key={heading.id}
          style={{
            paddingLeft: Math.min(heading.level - minimumLevel, 2) * 10,
          }}
        >
          <button
            aria-current={active === heading.element ? "location" : undefined}
            className={cn(
              "group hover:text-foreground focus-visible:ring-ring grid w-full grid-cols-[12px_minmax(0,1fr)] items-start gap-2 rounded py-1 text-left outline-none focus-visible:ring-2",
              active === heading.element
                ? "text-foreground"
                : "text-text-muted",
            )}
            onClick={() => {
              navigate(heading);
            }}
            title={heading.text}
            type="button"
          >
            <span
              aria-hidden
              className={cn(
                "mt-[0.6em] h-px w-2.5 origin-left bg-current transition-transform motion-reduce:transition-none",
                active === heading.element ? "scale-x-125" : "opacity-40",
              )}
            />
            <span className="line-clamp-2 leading-snug">{heading.text}</span>
          </button>
        </li>
      ))}
    </ol>
  );

  return (
    <div data-document-index>
      <nav aria-label="Document index" className={styles.desktop}>
        {links}
      </nav>
      <div
        className={styles.compact}
        style={{
          bottom: `calc(${compactPosition.bottom}px + env(safe-area-inset-bottom))`,
          left: compactPosition.left,
        }}
      >
        <Popover onOpenChange={setOpen} open={open}>
          <Popover.Trigger asChild>
            <Button
              aria-label="Open document index"
              asIcon
              className="border-border/80 bg-surface-elevated/85! shadow-shadow dark:bg-surface-elevated/85! size-10 shadow-lg backdrop-blur-xl"
              color="tertiary"
              rounded="full"
              variant="outline"
            >
              <ListIcon className="size-5" />
            </Button>
          </Popover.Trigger>
          <Popover.Content
            align="start"
            className="max-h-[min(56dvh,28rem)] w-72 max-w-[calc(100vw-2rem)] overflow-y-auto p-4"
            side="top"
            sideOffset={12}
          >
            <Text className="mb-3" fontWeight="medium">
              On this page
            </Text>
            <nav aria-label="Document index">{links}</nav>
          </Popover.Content>
        </Popover>
      </div>
    </div>
  );
}
