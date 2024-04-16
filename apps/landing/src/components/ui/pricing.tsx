"use client";
import { Flex, Text, Box, Button, Badge } from "ui";
import { ArrowRightIcon, ChatIcon, CheckIcon } from "icons";
import { Blur } from "./blur";
import { Container } from "./container";
import { cn } from "lib";
import { useState } from "react";
import { motion } from "framer-motion";

type Billing = "annual" | "monthly";
const packages = [
  {
    name: "Hobby",
    description: "Track your personal objectives",
    cta: "Start for free",
    overview:
      "Start your Hobby projects with our free plan. No credit card required.",
    price: 0,
    features: [
      "Use Kanban boards & lists",
      "All Integrations & API",
      "Google OAuth login",
      "Import & Export",
      "200 stories",
      "Importers",
    ],
  },
  {
    name: "Professional",
    description: "For small teams",
    cta: "Get started now",
    overview: "Everything in Hobby, plus more features for small teams.",
    price: 9,
    features: [
      "See product vision with Roadmaps",
      "Up tp 5 custom workflows",
      "Track OKRs",
      "Roadmaps",
      "Analytics & Reporting",
      "Custom fields",
    ],
  },
  {
    name: "Business",
    description: "For mid-sized teams",
    cta: "Get started now",
    overview:
      "Everything in Professional, plus more features for mid-sized teams.",
    price: 12,
    features: [
      "Unlimited everything",
      "Unlimited custom workflows",
      "Single Sign-On (SSO)",
      "Priority support",
      "Discussions & Whiteboards",
      "Advanced security",
    ],
    recommended: true,
  },
];

const Feature = ({ feature }: { feature: string }) => (
  <Flex align="center" gap={3} key={feature}>
    <Box className="flex aspect-square h-5 items-center justify-center rounded-full bg-gray-200">
      <CheckIcon strokeWidth={3.5} className="h-4 w-auto text-dark" />
    </Box>
    <Text className="opacity-90">{feature}</Text>
  </Flex>
);

const Package = ({
  name,
  description,
  overview,
  cta,
  price,
  features,
  recommended,
  billing,
}: {
  name: string;
  description: string;
  overview: string;
  cta: string;
  price: number;
  features: string[];
  recommended?: boolean;
  billing: Billing;
}) => {
  const [isActive, setIsActive] = useState(false);
  // if billing is annual, apply 25% discount
  if (billing === "annual") {
    price = price * 0.75;
  }
  return (
    <Box
      onMouseEnter={() => {
        setIsActive(true);
      }}
      onMouseLeave={() => {
        setIsActive(false);
      }}
      className="min-h-[65vh]"
    >
      <motion.div
        animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
        initial={{ y: 0, x: 0 }}
        transition={{ type: "spring", stiffness: 100 }}
        className={cn(
          "h-full rounded-3xl border-2 border-gray-200 bg-gray-50 p-8 shadow-2xl dark:border-dark-100 dark:bg-dark",
          {
            "border-primary shadow-primary/20 dark:border-primary": recommended,
          },
        )}
      >
        <Text
          className="mb-2 flex items-center justify-center gap-1.5 text-2xl"
          fontWeight="medium"
        >
          {name} {recommended && <Badge>Most Popular</Badge>}
        </Text>
        <Text className="text-center text-lg opacity-80" fontWeight="normal">
          {description}
        </Text>
        <Text
          align="center"
          fontSize="4xl"
          className="mt-4"
          fontWeight="medium"
        >
          ${price}
          <Text as="span" color="muted" fontSize="lg">
            /mo
          </Text>
        </Text>
        <Button
          className="mt-6 w-full justify-between"
          color="primary"
          rightIcon={<ArrowRightIcon className="h-4 w-auto" />}
          rounded="full"
          size="lg"
        >
          {cta}
        </Button>
        <Text className="my-6" color="muted">
          {overview}
        </Text>
        <Flex className="gap-4 md:gap-5" direction="column">
          {features.map((feature) => (
            <Feature feature={feature} key={feature} />
          ))}
        </Flex>
      </motion.div>
    </Box>
  );
};

export const Pricing = () => {
  const [billing, setBilling] = useState<Billing>("annual");
  const enterprise = [
    "Volume discounts",
    "Dedicated support",
    "Custom integrations",
    "Onboarding assistance",
    "Custom deployment",
    "SLA",
  ];

  return (
    <Box className="relative mb-20 md:mb-40">
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 z-[2] h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
      <Container className="md:pt-16">
        <Flex
          align="center"
          className="mb-8 mt-12 text-center md:mt-20"
          direction="column"
        >
          <Button
            className="px-3 text-sm md:text-base"
            color="tertiary"
            rounded="full"
            size="sm"
          >
            Pricing
          </Button>
          <Text
            as="h1"
            className="mt-6 h-max max-w-5xl pb-2 text-5xl md:text-7xl"
            color="gradient"
            fontWeight="medium"
          >
            Experience more, spend less. Switch to complexus.
          </Text>
          <Box className="mt-6">
            <Flex className="gap-1 rounded-[0.6rem] bg-dark-200 p-1">
              {["monthly", "annual"].map((option) => (
                <Button
                  key={option}
                  className={cn("px-2.5 capitalize", {
                    "opacity-80": option !== billing,
                  })}
                  color={option === billing ? "primary" : "tertiary"}
                  size="sm"
                  variant={option === billing ? "solid" : "naked"}
                  onClick={() => setBilling(option as Billing)}
                >
                  {option} Billing
                </Button>
              ))}
            </Flex>
            <Text className="mt-3">
              <Text as="span" color="primary" fontWeight="semibold">
                Save 25%
              </Text>{" "}
              with annual billing 🎉
            </Text>
          </Box>
        </Flex>

        <Box className="grid grid-cols-1 gap-8 md:grid-cols-3">
          {packages.map((pkg) => (
            <Package
              key={pkg.name}
              billing={billing}
              cta={pkg.cta}
              name={pkg.name}
              description={pkg.description}
              overview={pkg.overview}
              price={pkg.price}
              features={pkg.features}
              recommended={pkg.recommended}
            />
          ))}
        </Box>
        <Box className="relative mt-8 md:mt-32">
          <Box className="mx-auto rounded-3xl border-2 border-gray-200 bg-gray-50 p-8 shadow-2xl shadow-warning/10 dark:border-dark-100 dark:bg-dark md:max-w-4xl">
            <Text className="mb-3 text-2xl" fontWeight="medium">
              <span className="font-semibold text-primary">Complexus</span>{" "}
              Enteprise
            </Text>
            <Text className="text-lg opacity-80" fontWeight="normal">
              Crafted for enterprises aiming for seamless expansion. Complexus
              Enterprise provides state-of-the-art security measures, robust
              administrative capabilities, and enhanced features.
            </Text>

            <Box className="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2">
              {enterprise.map((feature) => (
                <Feature feature={feature} key={feature} />
              ))}
            </Box>
            <Button
              className="mt-6 justify-between"
              color="primary"
              leftIcon={<ChatIcon className="h-5 w-auto" />}
              rounded="full"
              size="lg"
            >
              Contact sales
            </Button>
          </Box>

          <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[500px] w-[500px] -translate-x-1/2 -translate-y-1/2 bg-warning/10" />
        </Box>
      </Container>
    </Box>
  );
};
