"use client";
import { Flex, Text, Box, Button, Badge, Divider } from "ui";
import { motion } from "framer-motion";
import { cn } from "lib";
import { useState } from "react";
import { CheckIcon } from "icons";
import { usePathname } from "next/navigation";
import { Container } from "./container";

type Billing = "annual" | "monthly";
const packages = [
  {
    name: "Hobby",
    cta: "Start for free",
    href: "/signup",
    overview: "Plan and track personal work with essential features.",
    price: 0,
    features: [
      "1 team",
      "Up to 5 members",
      "Up to 200 stories",
      "Single Sign-On (SSO)",
      "Kanban & list views",
      "Email support",
    ],
  },
  {
    name: "Professional",
    cta: "Try Professional",
    href: "/signup",
    overview: "Everything in Hobby, plus collaboration and OKR tracking.",
    price: 7,
    features: [
      "Up to 3 teams",
      "Up to 20 objectives",
      "Track OKRs",
      "Unlimited stories",
      "Unlimited guests",
      "Custom workflows",
    ],
  },
  {
    name: "Business",
    cta: "Try Business",
    href: "/signup",
    overview: "Everything in Professional, with advanced controls and support.",
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
    cta: "Contact sales",
    href: "mailto:info@complexus.app",
    overview: "Tailored deployment and support for complex environments.",
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
  <Flex align="center" className="gap-1.5" key={feature}>
    <CheckIcon />
    <Text className="text-[0.95rem] opacity-90">{feature}</Text>
  </Flex>
);

const Package = ({
  name,
  overview,
  price,
  features,
  recommended,
  billing,
  cta,
  href,
}: {
  name: string;
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
    <Box
      className={cn(
        "h-full rounded-3xl border border-gray-200 px-6 pb-8 pt-6 shadow-2xl shadow-gray-200 dark:border-dark-50 dark:bg-dark dark:shadow-dark",
        {
          "border-2 border-dark dark:border-white": recommended,
        },
      )}
    >
      <Text className="mb-2 flex items-center gap-1.5 text-xl font-semibold">
        {name}{" "}
        {recommended ? (
          <Badge color="invert" rounded="full">
            Most Popular
          </Badge>
        ) : null}
      </Text>

      {name !== "Enterprise" ? (
        <Text className="mt-4 text-4xl font-semibold">
          ${finalPrice % 1 === 0 ? finalPrice : finalPrice.toFixed(2)}
          <Text as="span" className="text-base opacity-60">
            {" "}
            {finalPrice > 0 ? "user/month" : ""}
          </Text>
        </Text>
      ) : (
        <Text className="mt-4" fontSize="4xl" fontWeight="semibold">
          Custom
        </Text>
      )}

      <Button
        align="center"
        className="mt-6 border-dark/30"
        color={recommended ? "invert" : "tertiary"}
        fullWidth
        href={href}
        variant={recommended ? "solid" : "outline"}
      >
        {cta}
      </Button>
      <Text className="mt-4">{overview}</Text>
      <Divider className="mb-5 mt-6 border-gray-200" />
      <Flex className="gap-4" direction="column">
        {features.map((feature) => (
          <Feature feature={feature} key={feature} />
        ))}
      </Flex>
    </Box>
  );
};

export const Pricing = ({ className }: { className?: string }) => {
  const pathname = usePathname();
  const [billing, setBilling] = useState<Billing>("annual");

  return (
    <Box className={cn("relative md:pt-12", className)}>
      <Container className="max-w-[85rem]">
        <Flex
          align="center"
          className={cn("my-12 text-center", {
            "md:mt-20": pathname === "/pricing",
          })}
          direction="column"
        >
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
              as={pathname === "/pricing" ? "h1" : "h2"}
              className="mt-6 max-w-3xl pb-2 text-5xl font-semibold md:text-6xl"
            >
              Start for free. Get used to hitting your{" "}
              <Text as="span" className="text-stroke-white">
                goals.
              </Text>
            </Text>
          </motion.div>
          <Text className="mt-4 max-w-2xl text-center text-xl opacity-70">
            Choose a plan that fits your needs with transparent pricing - no
            hidden fees, no unexpected charges, just clear value.
          </Text>
        </Flex>

        <Flex align="center" className="my-10" direction="column">
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.3,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Box className="flex w-max gap-1 rounded-full border border-gray-100 bg-white p-1 dark:border-dark-100 dark:bg-dark-300">
              {["annual", "monthly"].map((option) => (
                <Button
                  className={cn("px-3 capitalize", {
                    "opacity-80": option !== billing,
                  })}
                  color={option === billing ? "invert" : "tertiary"}
                  key={option}
                  onClick={() => {
                    setBilling(option as Billing);
                  }}
                  rounded="full"
                  size="sm"
                  variant={option === billing ? "solid" : "naked"}
                >
                  {option} Billing
                </Button>
              ))}
            </Box>
          </motion.div>
        </Flex>
        <Box className="grid grid-cols-1 gap-5 md:grid-cols-4">
          {packages.map((pkg) => (
            <Package
              billing={billing}
              cta={pkg.cta}
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
