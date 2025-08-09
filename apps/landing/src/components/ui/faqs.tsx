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
  <Box className="border-b border-gray-100 last:border-b-0 dark:border-dark-200">
    <button
      className={cn(
        "group flex w-full items-center justify-between py-6 text-left text-2xl outline-none",
      )}
      onClick={onToggle}
      type="button"
    >
      {item.question}
      <ArrowRight2Icon
        className={cn("h-6 shrink-0 transition-transform duration-300", {
          "rotate-90": isOpen,
        })}
        strokeWidth={2}
      />
    </button>
    <Box
      className={cn(
        "grid grid-rows-[0fr] transition-all duration-300 ease-in-out",
        {
          "grid-rows-1": isOpen,
        },
      )}
    >
      <Box className="overflow-hidden">
        <Text className="mb-10 max-w-2xl text-lg opacity-70">
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
    <Box className="py-16 md:pb-20 md:pt-24">
      <Container>
        <Text className="mb-12 text-5xl font-semibold leading-[1.1] md:text-5xl">
          Frequently Asked Questions
        </Text>
        <Flex className="pb-4" direction="column">
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
