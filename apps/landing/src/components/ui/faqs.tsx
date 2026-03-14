"use client";

import { Box, Flex, Text } from "ui";
import { ArrowRight2Icon, MinusIcon, PlusIcon } from "icons";
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
      "Maya is your AI project manager — not a chatbot bolted onto a to-do list. She drafts tasks from plain language, proposes sprint scope from your backlog, writes OKRs tied to your roadmap, and flags blockers before they cost you a sprint. She doesn't replace your team's judgment — she handles the coordination work so your team can apply it to better things. The longer she works with your setup, the more useful she gets.",
  },
  {
    question: "How does FortyOne connect goals to daily work?",
    answer:
      "Goals aren't a separate module — they're built into the structure of every sprint and task. Link work directly to key results and watch objective progress update automatically as tasks close. Your team sees exactly how their work moves the quarter. Leadership gets a live, trustworthy view without sending a single \"can you send me an update?\" email.",
  },
  {
    question: "Is the free plan actually free?",
    answer:
      "Yes — no credit card, no trial expiry, no watered-down version. The Hobby plan supports one team and up to five members, which is enough to run a real sprint and decide whether FortyOne is worth scaling. When you're ready to grow, paid plans start at $5.60 per user per month, billed annually.",
  },
  {
    question: "How does FortyOne handle security?",
    answer:
      "Encryption in transit and at rest, SSO with Google, role-based permissions, and private teams. For organizations with stricter requirements — compliance, on-premise, custom data residency — the Enterprise plan covers it with dedicated onboarding and a named account manager.",
  },
  {
    question: "Can we make it fit the way our team works?",
    answer:
      "That's the point. Customize statuses, workflows, terminology, and permissions to match your org — not a generic process template someone wrote in 2018. Structure your backlog and boards around how you plan and ship, and adjust as you grow without breaking historical context.",
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
            className="text-foreground dark:text-text-secondary h-6 shrink-0 transition-transform duration-300"
            strokeWidth={2}
          />
        ) : (
          <PlusIcon
            className="text-foreground dark:text-text-secondary h-6 shrink-0 transition-transform duration-300"
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
          <Text className="mb-10 max-w-6xl text-lg leading-relaxed opacity-60">
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
      <Container>
        <motion.div
          initial="hidden"
          variants={fadeUp}
          viewport={viewport}
          whileInView="show"
        >
          <Text
            as="h2"
            className="mb-6 text-5xl font-semibold md:mb-12 md:text-6xl"
          >
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
