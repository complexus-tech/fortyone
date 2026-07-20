"use client";

import { Box, Flex, Text } from "ui";
import { MinusIcon, PlusIcon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import { Container } from "./container";

type FaqItem = {
  question: string;
  answer: string;
};

const faqItems: FaqItem[] = [
  {
    question: "What makes FortyOne different from a task manager?",
    answer:
      "FortyOne connects customer feedback and company goals to the tasks that deliver them. Teams can decide what to build without losing the original request or the reason behind the work.",
  },
  {
    question: "How does customer feedback work?",
    answer:
      "Customers can submit requests, vote, and comment in a public portal. Feedback boards route each request to the team that owns the work. The team can review it, close it, or turn it into planned work.",
  },
  {
    question: "What can FortyOne's AI do?",
    answer:
      "AI can suggest an owner, fill in an estimate, find time for the work, and surface delivery risk from the project context already in FortyOne.",
  },
  {
    question: "Can we review AI actions before they are applied?",
    answer:
      "Yes. Teams can review and edit important suggestions before they change the project plan.",
  },
  {
    question: "Is the free plan actually free?",
    answer:
      "Yes. There is no credit card and no trial expiry. The Hobby plan supports one team and up to five members, enough to run a real sprint and decide whether FortyOne should scale with you.",
  },
  {
    question: "Can FortyOne plan around my team's calendar?",
    answer:
      "Yes. Google Calendar integration lets FortyOne sync availability so AI can recommend better schedules and work windows without storing private event details unnecessarily.",
  },
  {
    question: "Which tools does FortyOne work with?",
    answer:
      "FortyOne works with tools including Google Calendar, Slack, GitHub, GitLab, Figma, and Google Drive. These connections bring availability and source context into the project plan.",
  },
];

const AccordionItem = ({
  item,
  isOpen,
  onToggle,
  index,
}: {
  item: FaqItem;
  isOpen: boolean;
  onToggle: () => void;
  index: number;
}) => {
  const buttonId = `faq-trigger-${index}`;
  const panelId = `faq-panel-${index}`;
  return (
    <Box
      className="border-border d border-b last:border-b-0"
      data-landing-reveal
      style={{ transitionDelay: `${index * 60}ms` }}
    >
      <button
        aria-controls={panelId}
        aria-expanded={isOpen}
        className={cn(
          "group font-body focus-visible:ring-ring flex w-full items-start justify-between gap-6 py-5 text-left text-lg font-medium opacity-90 transition-[opacity,transform] duration-150 ease-out outline-none hover:opacity-100 focus-visible:ring-2 focus-visible:ring-inset active:scale-[0.995] motion-reduce:transition-none md:text-xl",
        )}
        id={buttonId}
        onClick={onToggle}
        type="button"
      >
        <span>{item.question}</span>
        <span aria-hidden="true" className="relative mt-0.5 size-5 shrink-0">
          <MinusIcon
            className={cn(
              "text-foreground dark:text-text-secondary absolute inset-0 transition-[opacity,transform] duration-200 ease-[cubic-bezier(0.23,1,0.32,1)] motion-reduce:transition-none",
              isOpen ? "rotate-0 opacity-100" : "-rotate-90 opacity-0",
            )}
            strokeWidth={2}
          />
          <PlusIcon
            className={cn(
              "text-foreground dark:text-text-secondary absolute inset-0 transition-[opacity,transform] duration-200 ease-[cubic-bezier(0.23,1,0.32,1)] motion-reduce:transition-none",
              isOpen ? "rotate-90 opacity-0" : "rotate-0 opacity-100",
            )}
            strokeWidth={2}
          />
        </span>
      </button>
      <Box
        aria-hidden={!isOpen}
        aria-labelledby={buttonId}
        className={cn(
          "grid transition-[grid-template-rows] duration-250 ease-[cubic-bezier(0.23,1,0.32,1)] motion-reduce:transition-none",
          isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
        )}
        id={panelId}
        role="region"
      >
        <Box
          className={cn(
            "overflow-hidden transition-[opacity,transform] duration-200 ease-[cubic-bezier(0.23,1,0.32,1)] motion-reduce:transform-none motion-reduce:transition-none",
            isOpen
              ? "translate-y-0 opacity-100 delay-50"
              : "-translate-y-1 opacity-0",
          )}
        >
          <Text className="max-w-2xl pb-7 text-base leading-relaxed opacity-60">
            {item.answer}
          </Text>
        </Box>
      </Box>
    </Box>
  );
};

export const Faqs = () => {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  const handleToggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <Box className="py-16 md:pt-24">
      <Container className="grid grid-cols-1 gap-8 md:grid-cols-[auto_1fr] md:justify-between md:gap-16">
        <Box data-landing-reveal>
          <Text as="h2" className="mb-2 text-4xl md:mb-12 md:text-5xl">
            Questions worth <br /> answering.
          </Text>
        </Box>

        <Flex
          className="w-full max-w-2xl justify-self-end pb-4"
          direction="column"
        >
          {faqItems.map((item, index) => (
            <AccordionItem
              index={index}
              isOpen={openIndex === index}
              item={item}
              key={item.question}
              onToggle={() => {
                handleToggle(index);
              }}
            />
          ))}
        </Flex>
      </Container>
    </Box>
  );
};
