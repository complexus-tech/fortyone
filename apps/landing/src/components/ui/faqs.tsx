"use client";

import { Box, Flex, Text } from "ui";
import { MinusIcon, PlusIcon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import type { HomeFaq } from "@/lib/home-faqs";
import { pricingFaqs } from "@/lib/home-faqs";
import { Container } from "./container";
import styles from "./faqs.module.css";

const AccordionItem = ({
  item,
  isOpen,
  onToggle,
  index,
  pricingStyle = false,
}: {
  item: HomeFaq;
  isOpen: boolean;
  onToggle: () => void;
  index: number;
  pricingStyle?: boolean;
}) => {
  const buttonId = `faq-trigger-${index}`;
  const panelId = `faq-panel-${index}`;
  return (
    <Box
      className="border-border border-b last:border-b-0"
      data-landing-reveal
      style={{ transitionDelay: `${index * 60}ms` }}
    >
      <button
        aria-controls={panelId}
        aria-expanded={isOpen}
        className={cn(
          "group font-body focus-visible:ring-ring flex w-full items-start justify-between gap-6 py-5 text-left font-medium opacity-90 transition-[opacity,transform] duration-150 ease-out outline-none hover:opacity-100 focus-visible:ring-2 focus-visible:ring-inset active:scale-[0.995] motion-reduce:transition-none md:text-lg",
          pricingStyle &&
            "justify-start gap-4 py-6 text-base font-semibold opacity-100 md:text-lg",
        )}
        id={buttonId}
        onClick={onToggle}
        type="button"
      >
        <span
          aria-hidden="true"
          className={cn(
            "relative mt-0.5 size-5 shrink-0",
            pricingStyle ? "order-first" : "order-last",
          )}
        >
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
        <span>{item.question}</span>
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
          <Text
            className={cn(
              "max-w-2xl pb-7 text-base leading-relaxed opacity-60",
              pricingStyle && "ml-9 opacity-70",
            )}
          >
            {item.answer}
          </Text>
        </Box>
      </Box>
    </Box>
  );
};

export const Faqs = ({
  variant = "default",
}: {
  variant?: "default" | "pricing";
}) => {
  const [openIndex, setOpenIndex] = useState<number | null>(0);
  const isPricing = variant === "pricing";

  const handleToggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <Box
      className={cn(
        "py-16 md:pt-24",
        isPricing &&
          "landing-hero-shell landing-page-frame mt-16 overflow-hidden rounded-[2.25rem] py-20 md:mt-20 md:rounded-[3rem] md:py-28",
        isPricing && styles.pricingShell,
      )}
    >
      <Container
        className={cn(
          "grid grid-cols-1 gap-8 md:grid-cols-[auto_1fr] md:justify-between md:gap-16",
          isPricing && "max-w-5xl gap-10 md:grid-cols-1 md:gap-12",
        )}
      >
        <Box className={cn(isPricing && "text-center")} data-landing-reveal>
          <Text
            as="h2"
            className={cn(
              "mb-2 text-4xl md:mb-12 md:text-5xl",
              isPricing &&
                "font-serif leading-tight font-medium tracking-[-0.035em] md:mb-0",
            )}
          >
            {isPricing ? (
              "Frequently asked questions"
            ) : (
              <>
                What teams ask <br aria-hidden="true" />
                before getting started.
              </>
            )}
          </Text>
        </Box>

        <Flex
          className={cn(
            "w-full max-w-2xl justify-self-end pb-4",
            isPricing && "mx-auto justify-self-auto",
          )}
          direction="column"
        >
          {pricingFaqs.map((item, index) => (
            <AccordionItem
              index={index}
              isOpen={openIndex === index}
              item={item}
              key={item.question}
              onToggle={() => {
                handleToggle(index);
              }}
              pricingStyle={isPricing}
            />
          ))}
        </Flex>
      </Container>
    </Box>
  );
};
