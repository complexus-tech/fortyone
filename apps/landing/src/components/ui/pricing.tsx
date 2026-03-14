"use client";
import { Flex, Text, Box, Button, Badge, Divider } from "ui";
import { motion } from "framer-motion";
import { cn } from "lib";
import { useState } from "react";
import { CheckIcon } from "icons";
import { usePathname } from "next/navigation";
import { SIGNUP_URL } from "@/lib/app-url";
import { Container } from "./container";

type Billing = "annual" | "monthly";
const packages = [
  {
    name: "Hobby",
    cta: "Start for free - no card needed",
    href: SIGNUP_URL,
    overview: "For individuals and small teams getting started.",
    price: 0,
    features: [
      "1 team",
      "Up to 5 members",
      "Up to 200 tasks",
      "Single Sign-On (SSO)",
      "Kanban & list views",
      "Email support",
    ],
  },
  {
    name: "Professional",
    cta: "Try Professional",
    href: SIGNUP_URL,
    overview: "For growing teams who need OKRs and more room to move.",
    price: 7,
    features: [
      "Everything in Hobby",
      "Up to 3 teams",
      "Up to 20 objectives",
      "OKR tracking",
      "Unlimited tasks",
      "Unlimited guests",
      "Custom workflows",
    ],
  },
  {
    name: "Business",
    cta: "Try Business",
    href: SIGNUP_URL,
    overview: "For teams who need full control, no limits, and fast support.",
    price: 10,
    features: [
      "Everything in Professional",
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
    cta: "Talk to sales",
    href: "mailto:info@complexus.app",
    overview:
      "For orgs with complex requirements, compliance needs, or on-premise preferences.",
    features: [
      "Custom onboarding & integrations",
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
        "border-border shadow-shadow h-full rounded-3xl border px-6 pt-6 pb-8 shadow-2xl",
        {
          "border-2 border-black dark:border-white": recommended,
        },
      )}
    >
      <Text className="mb-2 flex items-center gap-1.5 text-xl font-semibold">
        {name}{" "}
        {recommended ? (
          <Badge color="invert" className="font-semibold">
            Most Popular
          </Badge>
        ) : null}
      </Text>

      {name !== "Enterprise" ? (
        <Text className="mt-4 text-4xl font-bold">
          ${finalPrice % 1 === 0 ? finalPrice : finalPrice.toFixed(2)}
          <Text as="span" className="text-base font-medium opacity-60">
            {" "}
            {finalPrice > 0 ? "user/month" : ""}
          </Text>
        </Text>
      ) : (
        <Text className="mt-4" fontSize="4xl" fontWeight="bold">
          Custom
        </Text>
      )}

      <Button
        align="center"
        className={cn("border-border/80 mt-6", {
          "border-background-inverse": recommended,
        })}
        color={recommended ? "invert" : "tertiary"}
        fullWidth
        href={href}
        variant={recommended ? "solid" : "outline"}
      >
        {cta}
      </Button>
      <Text className="mt-4">{overview}</Text>
      <Divider className="border-border mt-6 mb-5" />
      <Flex className="gap-4" direction="column">
        {features.map((feature) => (
          <Feature feature={feature} key={feature} />
        ))}
      </Flex>
    </Box>
  );
};

export const Pricing = ({
  className,
  hideDescription,
}: {
  className?: string;
  hideDescription?: boolean;
}) => {
  const pathname = usePathname();
  const [billing, setBilling] = useState<Billing>("annual");

  return (
    <Box className={cn("relative md:pt-12", className)}>
      <Container className="max-w-332">
        <Box
          className={cn(
            "mt-10 flex flex-col gap-6 pb-6 md:flex-row md:items-end md:justify-between md:gap-16",
            {
            "md:mt-20": pathname === "/pricing",
            },
          )}
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
              className={cn("mt-6 max-w-3xl pb-2 text-4xl md:text-5xl")}
            >
              Start free. Scale when your team outgrows it.
            </Text>
          </motion.div>
          {!hideDescription ? (
            <Text className="w-full max-w-xl opacity-70 md:mt-4">
              No card required, no feature walls, no gotchas. The free plan
              handles a real team doing real work. Paid plans add more room and
              more Maya — upgrade, downgrade, or cancel any time.
            </Text>
          ) : null}
        </Box>

        <Flex className="mb-10" direction="column">
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.3,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Box className="border-border bg-surface flex w-max gap-1 rounded-full border p-1">
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
