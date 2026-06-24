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
    question: "What does Maya actually do?",
    answer:
      "Maya is the AI project manager inside FortyOne. She turns rough outcomes into tasks, proposes sprint scope from your backlog and capacity, connects work to goals, and flags blockers before they become end-of-sprint surprises.",
  },
  {
    question: "How does FortyOne connect goals to daily work?",
    answer:
      "Goals are not a reporting layer pasted on top of tasks. Work can be linked directly to objectives and key results, so progress updates as tasks move and leaders can see what changed without asking for another status report.",
  },
  {
    question: "Is the free plan actually free?",
    answer:
      "Yes. There is no credit card and no trial expiry. The Hobby plan supports one team and up to five members, enough to run a real sprint and decide whether FortyOne should scale with you.",
  },
  {
    question: "How does FortyOne handle security?",
    answer:
      "FortyOne supports encryption in transit and at rest, Google SSO, role-based permissions, and private teams. Enterprise teams can work with us on stricter requirements such as dedicated onboarding, private cloud, and deployment preferences.",
  },
  {
    question: "Can we make it fit the way our team works?",
    answer:
      "Yes. Customize statuses, workflows, terminology, permissions, teams, and planning rules around how you already ship. The structure can evolve as the team grows without throwing away the history behind prior work.",
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
          "group font-heading flex w-full items-start justify-between py-6 text-left text-xl opacity-90 outline-none md:text-2xl",
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
