"use client";
import { Flex, Text, Box, Button, Badge } from "ui";
import { ArrowRightIcon, ChatIcon, CheckIcon } from "icons";
import { Blur } from "./blur";
import { Container } from "./container";
import { cn } from "lib";
import { useState } from "react";
import { motion } from "framer-motion";

const packages = [
  {
    name: "Hobby",
    description: "Track your personal objectives",
    cta: "Start for free",
    overview:
      "Start your Hobby projects with our free plan. No credit card required.",
    price: 0,
    features: [
      "Up to 5 users",
      "Up to 5 guests",
      "1 team",
      "Unlimited stories",
      "Unlimited objectives",
      "Import & export data",
      "Basic support",
      "100MB storage",
    ],
  },
  {
    name: "Professional",
    description: "For small teams",
    cta: "Get started now",
    overview: "Everything in Hobby, plus more features for small teams.",
    price: 9,
    features: [
      "Unlimited users",
      "Unlimited guests",
      "Unlimited teams",
      "Integrations & API",
      "Priority support",
      "Advanced reporting",
      "50GB storage",
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
      "All Standard features",
      "Themes & Roadmaps",
      "Key Results",
      "Discussions",
      "100GB storage",
    ],
    recommended: true,
  },
];

const Feature = ({ feature }: { feature: string }) => (
  <Flex align="center" gap={2} key={feature}>
    <Box className="flex aspect-square h-5 items-center justify-center rounded-full bg-gray-200">
      <CheckIcon strokeWidth={3.5} className="h-4 w-auto text-primary" />
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
}: {
  name: string;
  description: string;
  overview: string;
  cta: string;
  price: number;
  features: string[];
  recommended?: boolean;
}) => {
  const [isActive, setIsActive] = useState(false);
  return (
    <Box
      onMouseEnter={() => {
        setIsActive(true);
      }}
      onMouseLeave={() => {
        setIsActive(false);
      }}
    >
      <motion.div
        animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
        initial={{ y: 0, x: 0 }}
        transition={{ type: "spring", stiffness: 100 }}
        className={cn(
          "h-full rounded-3xl border-2 border-dark-100 bg-dark p-8 shadow-2xl shadow-black",
          {
            "border-primary shadow-primary/20": recommended,
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
        <Text className="mt-6" color="muted">
          {overview}
        </Text>
        <Flex className="mt-4" direction="column" gap={4}>
          {features.map((feature) => (
            <Feature feature={feature} key={feature} />
          ))}
        </Flex>
      </motion.div>
    </Box>
  );
};

export const Pricing = () => {
  const enterprise = [
    "All Business features",
    "Volume discounts",
    "Dedicated support",
    "Custom integrations",
    "Onboarding assistance",
  ];

  return (
    <Box className="relative mb-40">
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 z-[2] h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
      <Container className="pt-16">
        <Flex
          align="center"
          className="my-16 text-center md:mt-20"
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
        </Flex>

        <Box className="grid grid-cols-1 gap-8 md:grid-cols-3">
          {packages.map((pkg) => (
            <Package
              key={pkg.name}
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
        <Box className="relative mt-32">
          <Box className="mx-auto rounded-3xl border-2 border-dark-100 bg-dark p-8 shadow-2xl shadow-warning/10 md:max-w-4xl">
            <Text className="mb-3 text-2xl" fontWeight="medium">
              <span className="font-semibold text-primary">Complexus</span>{" "}
              Enteprise
            </Text>
            <Text className="text-lg opacity-80" fontWeight="normal">
              Crafted for enterprises aiming for seamless expansion. Complexus
              Enterprise provides state-of-the-art security measures, robust
              administrative capabilities, and enhanced features.
            </Text>

            <Box className="mt-5 grid grid-cols-2 gap-4">
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

          <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[500px] w-[500px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
        </Box>
      </Container>
    </Box>
  );
};
