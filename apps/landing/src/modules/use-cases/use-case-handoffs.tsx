"use client";

import type { KeyboardEvent } from "react";
import { useRef, useState } from "react";
import { Text } from "ui";
import { Container } from "@/components/ui";
import type { UseCase } from "@/lib/use-cases";
import { ShowcaseHeading } from "@/modules/home/decide-what-matters-showcase";
import type { HandoffConfig } from "./handoff-types";
import { HandoffPreview } from "./handoff-preview";
import styles from "./handoffs.module.css";

export function UseCaseHandoffs({
  detail,
  config,
}: {
  detail: Pick<UseCase, "slug" | "label" | "sections">;
  config: HandoffConfig;
}) {
  const { workflows } = config;
  const { slug, sections } = detail;
  const [activeIndex, setActiveIndex] = useState(0);
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  function handleTabKeyDown(
    event: KeyboardEvent<HTMLButtonElement>,
    index: number,
  ) {
    let nextIndex: number;
    switch (event.key) {
      case "ArrowRight":
        nextIndex = (index + 1) % workflows.length;
        break;
      case "ArrowLeft":
        nextIndex = (index + workflows.length - 1) % workflows.length;
        break;
      case "Home":
        nextIndex = 0;
        break;
      case "End":
        nextIndex = workflows.length - 1;
        break;
      default:
        return;
    }
    event.preventDefault();
    setActiveIndex(nextIndex);
    tabRefs.current[nextIndex]?.focus();
  }

  return (
    <Container
      aria-labelledby={`${slug}-details-title`}
      as="section"
      className="scroll-mt-24 py-16 md:py-32"
      id={`${slug}-handoffs`}
    >
      <ShowcaseHeading
        description="The plan stays useful because the request, decision, owner, and delivery evidence remain connected as the work moves."
        id={`${slug}-details-title`}
        title="Keep the context behind every handoff."
        titleClassName="max-w-xl text-balance"
      />
      <div
        aria-label={`Explore ${detail.label.toLowerCase()} handoffs`}
        className={styles.tabs}
        role="tablist"
      >
        {workflows.map((workflow, index) => (
          <button
            aria-controls={`${slug}-handoff-panel-${workflow.id}`}
            aria-selected={activeIndex === index}
            className={styles.tab}
            id={`${slug}-handoff-tab-${workflow.id}`}
            key={workflow.id}
            onClick={() => {
              setActiveIndex(index);
            }}
            onKeyDown={(event) => {
              handleTabKeyDown(event, index);
            }}
            ref={(element) => {
              tabRefs.current[index] = element;
            }}
            role="tab"
            tabIndex={activeIndex === index ? 0 : -1}
            type="button"
          >
            <span aria-hidden="true" className={styles.tabNumber}>
              0{index + 1}
            </span>
            {workflow.label}
          </button>
        ))}
      </div>
      {workflows.map((workflow, index) => {
        const section = sections.find((item) => item.id === workflow.id);
        if (!section) {
          throw new Error(
            `Missing section ${workflow.id} for use case ${slug}`,
          );
        }
        return (
          <div
            aria-labelledby={`${slug}-handoff-tab-${workflow.id}`}
            className={styles.panel}
            hidden={activeIndex !== index}
            id={`${slug}-handoff-panel-${workflow.id}`}
            key={workflow.id}
            role="tabpanel"
            tabIndex={0}
          >
            <div className={styles.copy}>
              <Text
                as="h3"
                className="max-w-md text-2xl text-balance md:text-3xl"
              >
                {section.title}
              </Text>
              <div className="text-text-description mt-5 space-y-4 text-base leading-relaxed">
                {section.paragraphs.map((paragraph) => (
                  <p key={paragraph}>{paragraph}</p>
                ))}
              </div>
            </div>
            <figure className={styles.figure}>
              <HandoffPreview image={config.image} preview={workflow.preview} />
              <figcaption className="text-text-muted mt-5 text-center text-sm">
                {workflow.caption}
              </figcaption>
            </figure>
          </div>
        );
      })}
    </Container>
  );
}
