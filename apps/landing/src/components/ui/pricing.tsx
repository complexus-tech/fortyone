"use client";
import { Flex, Text, Box, Button, Badge, Divider } from "ui";
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
    cta: "Start free",
    href: SIGNUP_URL,
    overview: "For small teams running their first project.",
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
    overview: "For growing teams that need goals and more room to plan.",
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
    overview: "For organizations coordinating work across several teams.",
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
      "For organizations with security, compliance, deployment, or integration requirements.",
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
        "border-border bg-surface shadow-shadow h-full rounded-xl border px-6 pt-6 pb-8 shadow-lg",
        {
          "border-foreground border-2 shadow-xl": recommended,
        },
      )}
    >
      <Text className="mb-2 flex items-center gap-1.5 text-xl font-semibold">
        {name}{" "}
        {recommended ? (
          <Badge className="font-semibold" color="invert">
            Most Popular
          </Badge>
        ) : null}
      </Text>

      {name !== "Enterprise" ? (
        <Text className="mt-4 text-4xl font-semibold tracking-tight">
          ${finalPrice % 1 === 0 ? finalPrice : finalPrice.toFixed(2)}
          <Text as="span" className="text-base font-medium opacity-60">
            {" "}
            {finalPrice > 0 ? "user/month" : ""}
          </Text>
        </Text>
      ) : (
        <Text
          className="mt-4 tracking-tight"
          fontSize="4xl"
          fontWeight="semibold"
        >
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
      {/* Decorative colour blurs intentionally removed for the warmer background. */}
      <Container className="max-w-332">
        <Box
          className={cn(
            "mt-10 flex flex-col gap-6 pb-6 md:flex-row md:items-end md:justify-between md:gap-16",
            {
              "md:mt-20": pathname === "/pricing",
            },
          )}
        >
          <Box data-landing-reveal>
            <Text
              as={pathname === "/pricing" ? "h1" : "h2"}
              className={cn("mt-6 max-w-3xl pb-2 text-4xl md:text-5xl", {
                "text-5xl font-medium md:text-[3.5rem]":
                  pathname === "/pricing",
              })}
            >
              Start with one team. Scale when the work does.
            </Text>
          </Box>
          {!hideDescription ? (
            <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
              <Text className="w-full max-w-xl opacity-70 md:mt-4">
                No card and no trial clock. Run a real project on the free plan,
                then add teams, goals, integrations, and AI capacity as your
                organization grows.
              </Text>
            </Box>
          ) : null}
        </Box>

        <Flex className="mb-10" direction="column">
          <Box data-landing-reveal>
            <Box className="bg-surface-muted flex w-max gap-1 rounded-xl p-1">
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
                  rounded="md"
                  size="sm"
                  variant={option === billing ? "solid" : "naked"}
                >
                  {option} Billing
                </Button>
              ))}
            </Box>
          </Box>
        </Flex>
        <Box
          className="grid grid-cols-1 gap-5 md:grid-cols-4"
          data-landing-reveal
          style={{ transitionDelay: "70ms" }}
        >
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
