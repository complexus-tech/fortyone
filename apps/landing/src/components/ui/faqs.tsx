"use client";

import { Box, Flex, Text } from "ui";
import { ArrowRight2Icon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import { Container } from "./container";

type FaqItem = {
  question: string;
  answer: string;
};

const faqItems: FaqItem[] = [
  {
    question: "Can I use Complexus for free?",
    answer: "Absolutely! Start with our free tier. No credit card required.",
  },
  {
    question: "How many team members can I add on the free plan?",
    answer:
      "You can add up to 5 team members on our free plan, which is perfect just to get started.",
  },
  {
    question: "How flexible are your pricing plans?",
    answer:
      "Completely flexible. Upgrade when you need more features, downgrade when you don&apos;t, or cancel anytime with zero penalties or hidden fees. Your workspace, your terms.",
  },
  {
    question: "Are there special pricing options for startups or nonprofits?",
    answer:
      "Yes! We&apos;re passionate about supporting emerging teams and impactful organizations. Startups, educational institutions, and nonprofits can qualify for special discounts - just reach out to our team.",
  },
  {
    question: "What extra value do paid plans offer?",
    answer:
      "Paid plans unlock our power features: Objectives, OKRs, Permission controls, Priority support, and Custom workflows. Scale up as your team grows and your needs evolve.",
  },
  {
    question: "How quickly can my team get up and running?",
    answer:
      "Most teams are fully productive within minutes. Our intuitive interface requires minimal training, and we offer guided onboarding, templates, and documentation to accelerate your start.",
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
      <ArrowRight2Icon
        className={cn(
          "mt-1 h-5 shrink-0 transition-transform duration-300",
          isOpen && "rotate-90",
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
          <Text className="w-11/12 text-lg opacity-60">
            Everything you need to know about getting started with Complexus and
            our pricing plans.
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
