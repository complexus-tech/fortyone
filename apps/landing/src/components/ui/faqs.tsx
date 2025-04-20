"use client";

import { Box, Flex, Text } from "ui";
import { ArrowDownIcon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import { Container } from "./container";

type FaqItem = {
  question: string;
  answer: string;
};

const faqItems: FaqItem[] = [
  {
    question: "Is there a free plan for individuals or small teams?",
    answer:
      "Yes! Complexus offers a free tier that's perfect for individuals or small teams getting started. It includes core features like projects, sprints, objectives, and team collaboration—no credit card required.",
  },
  {
    question: "What happens when I reach the limits of the free plan?",
    answer:
      "We'll give you a heads-up before you hit any usage limits. You'll have the option to upgrade to a paid plan, but we never block access without notice or hold your data hostage.",
  },
  {
    question: "Can I change or cancel my plan anytime?",
    answer:
      "Absolutely. You can upgrade, downgrade, or cancel your plan whenever you like—no long-term commitments or hidden fees.",
  },
  {
    question: "Do you offer discounts for startups, students, or nonprofits?",
    answer:
      "Yes! We're startup-friendly and love supporting early-stage teams. Just reach out to us and we'll set you up with a discount if you qualify.",
  },
  {
    question: "What's the difference between free and paid plans?",
    answer:
      "Paid plans unlock advanced features like AI automation, OKR analytics, permission controls, and priority support. The free plan is great to start, and you can upgrade as your team grows.",
  },
];

const AccordionItem = ({
  item,
  isOpen,
  onToggle,
}: {
  item: FaqItem;
  isOpen: boolean;
  onToggle: () => void;
}) => (
  <Box className="rounded-2xl border border-gray-300 bg-white px-5 py-6 shadow-lg shadow-gray-200 dark:border-dark-200 dark:bg-dark/60 dark:shadow-none">
    <button
      className={cn(
        "group flex w-full justify-between text-left text-xl outline-none transition-all",
      )}
      onClick={onToggle}
      type="button"
    >
      {item.question}
      <ArrowDownIcon
        className={cn(
          "mt-1 h-5 shrink-0 transition-transform duration-300",
          isOpen && "rotate-180",
        )}
      />
    </button>
    <Box
      className={cn(
        "grid transition-all duration-300 ease-in-out",
        isOpen ? "mt-3 grid-rows-[1fr]" : "grid-rows-[0fr]",
      )}
    >
      <Box className="overflow-hidden">
        <Text className="text-base opacity-70">
          <span dangerouslySetInnerHTML={{ __html: item.answer }} />
        </Text>
      </Box>
    </Box>
  </Box>
);

export const Faqs = () => {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  const handleToggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <Box className="bg-dark/50 p-16 md:pb-20 md:pt-24">
      <Container className="grid grid-cols-1 gap-10 md:grid-cols-2 md:gap-16">
        <Box className="mx-auto max-w-3xl">
          <Text className="mb-6 text-4xl font-semibold leading-[1.1] md:text-[3.5rem]">
            Frequently Asked Questions
          </Text>
          <Text className="text-lg opacity-60">
            Here are the most common questions we receive, along with our
            answers.
          </Text>
        </Box>
        <Flex className="gap-4 pb-4" direction="column">
          {faqItems.map((item, index) => (
            <AccordionItem
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
