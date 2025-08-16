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
    question: "How does the AI assistant (Maya) help my team?",
    answer:
      "Maya accelerates planning and execution by drafting stories from plain text, proposing sprint scope from backlog context, and suggesting objectives and key results that map to your roadmap. During execution, Maya surfaces risks, highlights stuck work, and recommends follow ups based on ownership and recent activity.",
  },
  {
    question: "How does Complexus link OKRs to daily work?",
    answer:
      "Objectives and key results are first class. You can link stories and sprints directly to key results so progress rolls up automatically without spreadsheet wrangling or end of quarter scrambles. Contributors see exactly how their tasks drive outcomes, while leaders get a live, trustworthy view of progress and confidence levels.",
  },
  {
    question: "How is Complexus priced? Is there a free plan?",
    answer:
      "Yes. The Hobby plan is free with no credit card required and supports up to 1 team and 5 members to get started quickly. Paid plans are per user with monthly or annual billing, and annual saves 20%. Higher tiers unlock advanced workflows, custom terminology and permissions, unlimited teams, and priority support. You can upgrade, downgrade, or cancel at any time.",
  },
  {
    question:
      "How does Complexus handle security and privacy? Is private cloud available?",
    answer:
      "We use industry standard encryption in transit and at rest, support SSO with Google, and provide role based permissions and private teams to control access. Audit friendly activity history and fine grained visibility help maintain compliance practices. For organizations with stricter controls, the Enterprise plan offers private cloud or on premise deployment with tailored onboarding.",
  },
  {
    question: "Can Complexus adapt to our workflow?",
    answer:
      "Yes. Teams can customize statuses and workflows, define their own terminology, and set granular permissions by role or team. Automations help reduce repetitive work, and you can structure backlogs and boards to mirror how your org plans and executes. As needs evolve, you can adjust configurations without breaking historical data or reports.",
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
    <Box className="border-b border-gray-100 last:border-b-0 dark:border-dark-200">
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
            "h-6 shrink-0 text-dark transition-transform duration-300 dark:text-gray-200",
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
          <Text className="mb-10 max-w-3xl pl-1 text-lg opacity-60 md:text-xl">
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
            className="mb-6 text-4xl font-semibold leading-[1.1] md:mb-12 md:text-5xl"
          >
            Frequently Asked Questions
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
