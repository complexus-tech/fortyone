"use client";
import { Flex, Text, Box, Button, Badge } from "ui";
import { motion } from "framer-motion";
import { cn } from "lib";
import { useState } from "react";
import { SuccessIcon } from "icons";
import { Container } from "./container";

type Billing = "annual" | "monthly";
const packages = [
  {
    name: "Hobby",
    description: "Track your personal objectives",
    cta: "Start for free",
    href: "/signup",
    overview:
      "Start your Hobby projects with our free plan. No credit card required.",
    price: 0,
    features: [
      "1 team",
      "Up to 5 members",
      "Up to 50 stories",
      "Single Sign-On (SSO)",
      "Kanban & list views",
      "Email support",
    ],
  },
  {
    name: "Professional",
    description: "For small teams",
    cta: "Upgrade now",
    href: "/signup",
    overview: "Everything in Hobby, plus more features for small teams.",
    price: 7,
    features: [
      "Up to 3 teams",
      "Up to 5 objectives",
      "Track OKRs",
      "Unlimited stories",
      "Unlimited guests",
      "Custom workflows",
    ],
  },
  {
    name: "Business",
    description: "For mid-sized teams",
    cta: "Upgrade now",
    href: "/signup",
    overview:
      "Everything in Professional, plus more features for mid-sized teams.",
    price: 10,
    features: [
      "Unlimited teams",
      "Unlimited objectives",
      "Unlimited everything",
      "Custom terminology",
      "Private teams",
      "Priority support",
    ],
    recommended: true,
  },
  {
    name: "Enterprise",
    description: "For large teams",
    cta: "Contact sales",
    href: "mailto:info@complexus.app",
    overview: "Everything in Professional, plus more features for large teams.",
    features: [
      "Custom onboarding",
      "Custom integrations",
      "On-premise/Private Cloud Option",
      "Dedicated account manager",
      "Volume discounts",
    ],
  },
];

const Feature = ({ feature }: { feature: string }) => (
  <Flex align="center" gap={2} key={feature}>
    <SuccessIcon className="h-[1.35rem] dark:text-primary" />
    <Text className="opacity-90">{feature}</Text>
  </Flex>
);

const Package = ({
  name,
  description,
  overview,
  price,
  features,
  recommended,
  billing,
  cta,
  href,
}: {
  name: string;
  description: string;
  overview: string;
  cta: string;
  price?: number;
  features: string[];
  recommended?: boolean;
  billing: Billing;
  href: string;
}) => {
  // if billing is annual, apply 20% discount
  let finalPrice = price ?? 0;
  if (billing === "annual" && price) {
    finalPrice = price * 0.8;
  }
  return (
    <Box className="min-h-[65vh]">
      <Box
        className={cn("h-full px-6 py-8 shadow-2xl dark:bg-dark", {
          "border-2 border-primary bg-gray-50 shadow-primary/20 dark:bg-dark-300":
            recommended,
        })}
      >
        <Text className="mb-2 flex items-center justify-center gap-1.5 text-2xl">
          {name} {recommended ? <Badge>Most Popular</Badge> : null}
        </Text>
        <Text className="text-center text-lg opacity-80">{description}</Text>
        {name !== "Enterprise" ? (
          <Text align="center" className="mt-4" fontSize="4xl">
            ${finalPrice % 1 === 0 ? finalPrice : finalPrice.toFixed(2)}
            <Text as="span" color="muted" fontSize="lg">
              {" "}
              {finalPrice > 0 ? "per user/month" : ""}
            </Text>
          </Text>
        ) : (
          <Text align="center" className="mt-4" fontSize="4xl">
            Custom Pricing
          </Text>
        )}

        <Button
          align="center"
          className="mt-6 md:h-11"
          color={recommended ? "primary" : "tertiary"}
          fullWidth
          href={href}
          rounded="lg"
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
      </Box>
    </Box>
  );
};

export const Pricing = () => {
  const [billing, setBilling] = useState<Billing>("annual");

  return (
    <Box className="relative">
      <Container className="md:pt-16">
        <Flex
          align="center"
          className="my-12 text-center md:mt-20"
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
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text
              as="h1"
              className="mt-6 max-w-4xl pb-2 text-5xl font-semibold md:text-7xl"
            >
              Simple pricing for ambitious teams
            </Text>
          </motion.div>
          <Text className="mt-3 max-w-2xl text-center text-xl opacity-80">
            Choose a plan that fits your needs with transparent pricing - no
            hidden fees, no unexpected charges, just clear value.
          </Text>
        </Flex>

        <Flex align="end" className="mb-4 mt-12" direction="column">
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.3,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Box className="flex w-max gap-1 rounded-[0.7rem] border border-gray-100 bg-white p-1 dark:border-dark-100 dark:bg-dark-300">
              {["annual", "monthly"].map((option) => (
                <Button
                  className={cn("px-2.5 capitalize", {
                    "opacity-80": option !== billing,
                  })}
                  color={option === billing ? "primary" : "tertiary"}
                  key={option}
                  onClick={() => {
                    setBilling(option as Billing);
                  }}
                  size="sm"
                  variant={option === billing ? "solid" : "naked"}
                >
                  {option} Billing
                </Button>
              ))}
            </Box>
          </motion.div>
          <Box className="mt-2">
            <motion.p
              initial={{ y: 20, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.6,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Text as="span" color="primary" fontWeight="semibold">
                Save 20%
              </Text>{" "}
              with annual billing 🎉
            </motion.p>
          </Box>
        </Flex>
        <Box className="grid grid-cols-1 divide-y divide-dark-100 overflow-hidden rounded-3xl border border-dark-100 bg-white dark:bg-dark md:grid-cols-4 md:divide-x">
          {packages.map((pkg) => (
            <Package
              billing={billing}
              cta={pkg.cta}
              description={pkg.description}
              features={pkg.features}
              href={pkg.href}
              key={pkg.name}
              name={pkg.name}
              overview={pkg.overview}
              price={pkg.price}
              recommended={pkg.recommended}
            />
          ))}
        </Box>
      </Container>
    </Box>
  );
};
