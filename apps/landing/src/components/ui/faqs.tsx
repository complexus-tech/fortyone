"use client";

import { Box, Flex, Text } from "ui";
import { ArrowRight2Icon } from "icons";
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
      "Maya is your AI project manager - not just a chatbot bolted onto a to-do list. She drafts tasks from plain text, proposes sprint scope from your backlog, writes goals and key results that connect to your roadmap, and flags blockers in real time based on ownership and activity.",
  },
  {
    question: "How does FortyOne connect goals to daily work?",
    answer:
      "Goals are built into the structure of every sprint and task. Link work directly to key results and watch progress roll up automatically. Team members see how their work drives outcomes, and leaders get a live, trustworthy view of where things stand without spreadsheet wrangling.",
  },
  {
    question: "Is there really a free plan?",
    answer:
      "Yes. The Hobby plan is free with no credit card required and supports up to 1 team and 5 members, which is enough to get real work done and decide if FortyOne is right for you. Paid plans are per user, with annual billing saving 20%. You can upgrade, downgrade, or cancel at any time.",
  },
  {
    question: "How does FortyOne handle security?",
    answer:
      "We use industry-standard encryption in transit and at rest, support SSO with Google, and provide role-based permissions and private teams to control access. For organizations with stricter requirements, the Enterprise plan includes private cloud or on-premise deployment with tailored onboarding.",
  },
  {
    question: "Can we make it fit the way our team works?",
    answer:
      "Yes. Customize statuses, workflows, terminology, and permissions to match your org - not a generic process template. Structure your backlog and boards around the way your team plans and executes, and adjust things as you grow without breaking historical context.",
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
          "group flex w-full items-start justify-between py-6 text-left text-xl opacity-90 outline-none md:text-2xl",
        )}
        id={buttonId}
        onClick={onToggle}
        type="button"
      >
        {item.question}

        <ArrowRight2Icon
          className={cn(
            "text-foreground dark:text-text-secondary h-6 shrink-0 transition-transform duration-300",
            {
              "rotate-90": isOpen,
            },
          )}
          strokeWidth={2}
        />
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
          <Text className="mb-10 max-w-3xl text-lg leading-relaxed opacity-60">
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
      <Container className="max-w-4xl">
        <motion.div
          initial="hidden"
          variants={fadeUp}
          viewport={viewport}
          whileInView="show"
        >
          <Text
            as="h2"
            className="mb-6 text-4xl leading-[1.1] font-bold md:mb-12 md:text-5xl"
          >
            Questions worth answering.
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
