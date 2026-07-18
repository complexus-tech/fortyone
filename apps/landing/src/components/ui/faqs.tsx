"use client";

import { Box, Flex, Text } from "ui";
import { MinusIcon, PlusIcon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import { motion } from "framer-motion";
import { Container } from "./container";

const viewport = { once: true, amount: 0.35 };
const fadeUp = {
  hidden: { y: 16, opacity: 0 },
  show: { y: 0, opacity: 1, transition: { duration: 0.6, ease: "easeOut" } },
};

type FaqItem = {
  question: string;
  answer: string;
};

const faqItems: FaqItem[] = [
  {
    question: "What makes FortyOne an AI project manager?",
    answer:
      "FortyOne combines project management software with an AI assistant that can turn project context into tasks, suggest owners, add estimates, plan timing, and surface delivery risks before work slips.",
  },
  {
    question: "What happens when I assign work to AI?",
    answer:
      "The AI reviews the task, team context, workload, estimates, and availability, then helps find the right owner, schedule, and next action. Admins can review important AI actions before they are applied.",
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
    question: "Can we create tasks from Slack?",
    answer:
      "Yes. Slack support lets teams create tasks from Slack and ask the AI assistant for help where conversations already happen, while keeping the project plan in FortyOne.",
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
    <Box className="border-border d border-b last:border-b-0">
      <button
        aria-controls={panelId}
        aria-expanded={isOpen}
        className={cn(
          "group font-body flex w-full items-start justify-between py-6 text-left text-xl font-medium opacity-90 outline-none md:text-2xl",
        )}
        id={buttonId}
        onClick={onToggle}
        type="button"
      >
        {item.question}

        {isOpen ? (
          <MinusIcon
            className="text-foreground dark:text-text-secondary shrink-0 transition-transform duration-300"
            strokeWidth={2}
          />
        ) : (
          <PlusIcon
            className="text-foreground dark:text-text-secondary shrink-0 transition-transform duration-300"
            strokeWidth={2}
          />
        )}
      </button>
      <Box
        aria-labelledby={buttonId}
        className={cn(
          "grid grid-rows-[0fr] transition-all duration-300 ease-in-out",
          {
            "grid-rows-1": isOpen,
          },
        )}
        id={panelId}
        role="region"
      >
        <Box className="overflow-hidden">
          <Text className="mb-10 max-w-6xl text-base leading-relaxed opacity-60">
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
        <motion.div
          initial="hidden"
          variants={fadeUp}
          viewport={viewport}
          whileInView="show"
        >
          <Text as="h2" className="mb-2 text-4xl md:mb-12 md:text-5xl">
            Questions worth <br /> answering.
          </Text>
        </motion.div>

        <Flex className="pb-4" direction="column">
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
