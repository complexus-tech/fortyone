"use client";
import type { ComponentPropsWithoutRef, ComponentType } from "react";
import { CheckIcon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import { Badge, Box, Button, Divider, Flex, Switch, Text } from "ui";
import { SIGNUP_URL } from "@/lib/app-url";
import { Container } from "./container";
import styles from "./pricing.module.css";
import {
  PricingAccessIcon,
  PricingAdminIcon,
  PricingAiIcon,
  PricingArrowIcon,
  PricingDeploymentIcon,
  PricingFeedbackIcon,
  PricingIntegrationIcon,
  PricingOrganizationIcon,
  PricingPlanningIcon,
  PricingScaleIcon,
  PricingSupportIcon,
  PricingTeamIcon,
  PricingWorkIcon,
} from "./pricing-icons";

type Billing = "annual" | "monthly";
type PricingVariant = "page" | "section";
type FeatureIcon = ComponentType<ComponentPropsWithoutRef<"svg">>;

type FeatureGroupData = {
  title: string;
  icon: FeatureIcon;
  iconClassName: string;
  iconSurfaceClassName: string;
  features: string[];
};

const BILLING_OPTIONS: { label: string; value: Billing }[] = [
  { label: "Annual", value: "annual" },
  { label: "Monthly", value: "monthly" },
];

const PRICING_PAGE_BILLING_OPTIONS: { label: string; value: Billing }[] = [
  { label: "Billed monthly", value: "monthly" },
  { label: "Billed yearly", value: "annual" },
];

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
      "1 objective with OKR tracking",
      "Customer feedback & public roadmap",
      "15 Maya AI agent messages per month",
      "Single Sign-On (SSO)",
      "Kanban & list views",
      "Email support",
    ],
    featureGroups: [
      {
        title: "Work management",
        icon: PricingWorkIcon,
        iconClassName: "text-info",
        iconSurfaceClassName: "bg-info/15",
        features: [
          "1 team",
          "Up to 5 members",
          "Up to 200 tasks",
          "Kanban & list views",
          "Sprints, My Work & summaries",
          "Shared project documents",
        ],
      },
      {
        title: "Strategy & feedback",
        icon: PricingFeedbackIcon,
        iconClassName: "text-primary",
        iconSurfaceClassName: "bg-primary/12",
        features: [
          "1 objective with OKR tracking",
          "Customer feedback boards & voting",
          "Public feedback portal & roadmap",
        ],
      },
      {
        title: "Maya AI agent",
        icon: PricingAiIcon,
        iconClassName: "text-success",
        iconSurfaceClassName: "bg-success/15",
        features: [
          "15 AI agent messages per month",
          "Calendar and focus planning",
        ],
      },
      {
        title: "Integrations",
        icon: PricingIntegrationIcon,
        iconClassName: "text-secondary",
        iconSurfaceClassName: "bg-secondary/15",
        features: [
          "Slack work intake & Maya",
          "GitHub issue sync",
          "Figma & Google Calendar context",
        ],
      },
      {
        title: "Access & support",
        icon: PricingAccessIcon,
        iconClassName: "text-secondary",
        iconSurfaceClassName: "bg-secondary/15",
        features: ["Single Sign-On (SSO)", "Email support"],
      },
    ] satisfies FeatureGroupData[],
  },
  {
    name: "Professional",
    cta: "Try Professional",
    href: SIGNUP_URL,
    overview:
      "For teams that need shared goals, custom workflows, and room to plan across multiple teams.",
    price: 7,
    features: [
      "Everything in Hobby",
      "Up to 3 teams",
      "Up to 20 objectives",
      "OKR tracking",
      "Unlimited tasks",
      "Unlimited guests",
      "Custom workflows",
      "100 Maya AI agent messages per month",
      "AI work planning",
    ],
    featureGroups: [
      {
        title: "Team planning",
        icon: PricingTeamIcon,
        iconClassName: "text-info",
        iconSurfaceClassName: "bg-info/15",
        features: ["Up to 3 teams", "Unlimited tasks", "Unlimited guests"],
      },
      {
        title: "Goals & workflows",
        icon: PricingPlanningIcon,
        iconClassName: "text-primary",
        iconSurfaceClassName: "bg-primary/12",
        features: ["Up to 20 objectives", "OKR tracking", "Custom workflows"],
      },
      {
        title: "Maya AI agent",
        icon: PricingAiIcon,
        iconClassName: "text-success",
        iconSurfaceClassName: "bg-success/15",
        features: [
          "100 AI agent messages per month",
          "AI owner & work-window suggestions",
          "Calendar-aware task scheduling",
        ],
      },
    ] satisfies FeatureGroupData[],
  },
  {
    name: "Business",
    cta: "Try Business",
    href: SIGNUP_URL,
    overview:
      "For organizations coordinating multiple teams with private spaces, shared terminology, and no planning limits.",
    price: 10,
    features: [
      "Everything in Professional",
      "Unlimited teams",
      "Unlimited objectives",
      "Unlimited everything",
      "Custom terminology",
      "Private teams",
      "Priority support",
      "500 Maya AI agent messages per month",
    ],
    featureGroups: [
      {
        title: "Organization",
        icon: PricingOrganizationIcon,
        iconClassName: "text-info",
        iconSurfaceClassName: "bg-info/15",
        features: [
          "Unlimited teams",
          "Unlimited objectives",
          "Unlimited everything",
        ],
      },
      {
        title: "Administration",
        icon: PricingAdminIcon,
        iconClassName: "text-secondary",
        iconSurfaceClassName: "bg-secondary/15",
        features: ["Custom terminology", "Private teams"],
      },
      {
        title: "AI agent at scale",
        icon: PricingAiIcon,
        iconClassName: "text-primary",
        iconSurfaceClassName: "bg-primary/12",
        features: [
          "500 AI agent messages per month",
          "AI work planning across unlimited teams",
          "Automated planning actions",
        ],
      },
      {
        title: "Support",
        icon: PricingSupportIcon,
        iconClassName: "text-success",
        iconSurfaceClassName: "bg-success/15",
        features: ["Priority support"],
      },
    ] satisfies FeatureGroupData[],
    recommended: true,
  },
  {
    name: "Enterprise",
    cta: "Talk to sales",
    href: "mailto:info@complexus.app",
    overview:
      "For organizations with security, compliance, deployment, or integration requirements.",
    features: [
      "Everything in Business",
      "Custom onboarding & integrations",
      "On-premise/Private Cloud Option",
      "Dedicated account manager",
      "Volume discounts",
      "Unlimited Maya AI agent messages",
    ],
    featureGroups: [
      {
        title: "Deployment & security",
        icon: PricingDeploymentIcon,
        iconClassName: "text-secondary",
        iconSurfaceClassName: "bg-secondary/15",
        features: ["On-premise or private cloud deployment"],
      },
      {
        title: "Services & scale",
        icon: PricingScaleIcon,
        iconClassName: "text-success",
        iconSurfaceClassName: "bg-success/15",
        features: [
          "Custom onboarding & integrations",
          "Dedicated account manager",
          "Volume discounts",
        ],
      },
      {
        title: "AI agent & planning",
        icon: PricingAiIcon,
        iconClassName: "text-primary",
        iconSurfaceClassName: "bg-primary/12",
        features: ["Unlimited AI agent messages", "AI work planning at scale"],
      },
    ] satisfies FeatureGroupData[],
  },
];

const Feature = ({ feature }: { feature: string }) => (
  <Flex align="start" as="li" className="gap-2.5">
    <CheckIcon aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
    <Text className="text-[0.95rem] leading-6">{feature}</Text>
  </Flex>
);

const FeatureGroup = ({ group }: { group: FeatureGroupData }) => {
  const Icon = group.icon;

  return (
    <Box className="border-border border-t pt-5 first:border-t-0 first:pt-0">
      <Flex align="center" className="mb-4 gap-2.5">
        <Box
          className={cn(
            "flex size-7 shrink-0 items-center justify-center rounded-lg",
            group.iconSurfaceClassName,
          )}
        >
          <Icon
            aria-hidden="true"
            className={cn("size-[1.125rem]", group.iconClassName)}
            strokeWidth={1.7}
          />
        </Box>
        <Text className="text-base font-semibold">{group.title}</Text>
      </Flex>
      <Flex as="ul" className="gap-3" direction="column">
        {group.features.map((feature) => (
          <Feature feature={feature} key={feature} />
        ))}
      </Flex>
    </Box>
  );
};

const BillingSelector = ({
  billing,
  onChange,
  pricingPageStyle = false,
  showBillingSuffix = false,
  showSavings = false,
}: {
  billing: Billing;
  onChange: (billing: Billing) => void;
  pricingPageStyle?: boolean;
  showBillingSuffix?: boolean;
  showSavings?: boolean;
}) => {
  if (pricingPageStyle) {
    return (
      <>
        <Flex
          align="center"
          aria-label="Billing period"
          className="hidden flex-wrap gap-4 md:flex"
          role="radiogroup"
        >
          {PRICING_PAGE_BILLING_OPTIONS.map((option, index) => (
            <Flex align="center" className="gap-4" key={option.value}>
              {index > 0 ? (
                <Divider className="h-5 border-t-0 border-l" />
              ) : null}
              <Text
                as="label"
                className="flex cursor-pointer items-center gap-2 text-base font-normal"
              >
                <input
                  aria-label={option.label}
                  checked={billing === option.value}
                  className="accent-secondary size-5 cursor-pointer"
                  name="pricing-billing-period"
                  onChange={() => {
                    onChange(option.value);
                  }}
                  type="radio"
                  value={option.value}
                />
                {option.label}
              </Text>
            </Flex>
          ))}
          {showSavings ? (
            <Badge
              className="bg-state-selected text-foreground h-6 border-0 px-2.5"
              color="invert"
              rounded="md"
              size="sm"
            >
              Save up to 20%
            </Badge>
          ) : null}
        </Flex>
        <Text
          as="label"
          className="flex cursor-pointer items-center justify-center gap-2 text-sm font-normal md:hidden"
        >
          <Switch
            aria-label="Toggle yearly billing"
            checked={billing === "annual"}
            className="h-5 min-h-5 w-9 min-w-9 [&>span]:size-4 [&>span]:data-[state=checked]:translate-x-4"
            onCheckedChange={(checked) => {
              onChange(checked ? "annual" : "monthly");
            }}
          />
          {billing === "annual" ? "Billed yearly" : "Billed monthly"}
        </Text>
      </>
    );
  }

  return (
    <Flex align="center" className="flex-wrap gap-3">
      <Box
        aria-label="Billing period"
        className="bg-background/70 dark:bg-surface-elevated/80 flex w-max gap-1 rounded-xl p-1 shadow-sm"
        role="radiogroup"
      >
        {BILLING_OPTIONS.map((option) => {
          const isSelected = option.value === billing;

          return (
            <Button
              aria-checked={isSelected}
              className={cn("px-3", {
                "text-text-muted opacity-80": !isSelected,
              })}
              color={isSelected ? "invert" : "tertiary"}
              key={option.value}
              onClick={() => {
                onChange(option.value);
              }}
              role="radio"
              rounded="md"
              size="sm"
              variant={isSelected ? "solid" : "naked"}
            >
              {option.label}
              {showBillingSuffix ? " Billing" : null}
            </Button>
          );
        })}
      </Box>
      {showSavings ? (
        <Badge
          className="bg-primary/10 text-primary border-0 px-2.5"
          color="primary"
          rounded="full"
          size="sm"
        >
          Save 20%
        </Badge>
      ) : null}
    </Flex>
  );
};

const Package = ({
  name,
  overview,
  price,
  features,
  featureGroups,
  recommended,
  billing,
  cta,
  href,
  variant,
}: {
  name: string;
  overview: string;
  cta: string;
  price?: number;
  features: string[];
  featureGroups: FeatureGroupData[];
  recommended?: boolean;
  billing: Billing;
  href: string;
  variant: PricingVariant;
}) => {
  // if billing is annual, apply 20% discount
  let finalPrice = price ?? 0;
  if (billing === "annual" && price) {
    finalPrice = price * 0.8;
  }
  let displayPrice = "Custom";
  if (name !== "Enterprise") {
    displayPrice =
      finalPrice === 0
        ? "Always free"
        : `$${finalPrice % 1 === 0 ? finalPrice : finalPrice.toFixed(2)}`;
  }
  const inheritsPreviousPlan = features[0]?.startsWith("Everything in ");
  const featureHeading = inheritsPreviousPlan
    ? `${features[0]}, and:`
    : "Includes:";

  if (variant === "page") {
    return (
      <Box
        className={cn(
          "relative h-full",
          recommended &&
            `${styles.featured} -mt-10 h-[calc(100%+2.5rem)] rounded-[2.75rem] p-2 pt-10 md:rounded-[3.5rem]`,
        )}
      >
        {recommended ? (
          <Box className="absolute top-3 left-8 text-xs font-semibold tracking-[0.12em] text-[var(--brand-paper)] uppercase">
            Most popular
          </Box>
        ) : null}
        <Box className="border-border/60 bg-surface-muted dark:bg-surface-elevated shadow-shadow relative z-1 flex h-full flex-col gap-7 rounded-[2.5rem] border p-2 pb-7 shadow-xl md:min-h-[68rem] md:rounded-[3rem]">
          <Box className="border-border/60 bg-surface shadow-shadow dark:bg-surface flex min-h-[18rem] flex-col rounded-[2rem] border p-7 shadow-lg md:min-h-[19rem] md:rounded-[2.5rem] md:p-8">
            <Text className="text-3xl font-semibold tracking-[-0.035em] md:text-[2rem]">
              {name}
            </Text>
            <Text
              className={cn(
                "mt-auto font-serif text-[2.65rem] leading-none font-medium tracking-[-0.045em] whitespace-nowrap md:text-[2.85rem]",
                name === "Hobby" && "md:text-[2.15rem]",
              )}
            >
              {displayPrice}
              {name !== "Enterprise" && finalPrice > 0 ? (
                <Text
                  as="span"
                  className="text-text-muted ml-1 font-sans text-sm font-medium tracking-normal"
                >
                  /user/mo
                </Text>
              ) : null}
            </Text>
            <Button
              align="between"
              className="mt-4 px-4 whitespace-nowrap"
              color="invert"
              fullWidth
              href={href}
              rightIcon={
                <PricingArrowIcon
                  aria-hidden="true"
                  className="h-5 w-auto"
                  strokeWidth={2}
                />
              }
              rounded="md"
              size="lg"
              variant={recommended ? "solid" : "outline"}
            >
              {cta}
            </Button>
          </Box>
          <Box className="flex flex-1 flex-col px-4 md:px-5">
            <Text className="mb-5 text-sm leading-6">{featureHeading}</Text>
            <Text className="mb-5 text-sm leading-6 font-normal opacity-70">
              {overview}
            </Text>
            <Flex className="gap-5" direction="column">
              {featureGroups.map((group) => (
                <FeatureGroup group={group} key={group.title} />
              ))}
            </Flex>
          </Box>
        </Box>
      </Box>
    );
  }

  return (
    <Box
      className={cn(
        "border-border/80 bg-surface dark:bg-surface/60 shadow-shadow h-full rounded-xl border px-6 pt-6 pb-8 shadow-lg",
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

      <Text className="mt-4 text-4xl font-semibold tracking-tight">
        {displayPrice}
        {name !== "Enterprise" && finalPrice > 0 ? (
          <Text as="span" className="text-base font-medium opacity-60">
            {" "}
            user/month
          </Text>
        ) : null}
      </Text>

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
      <Flex as="ul" className="gap-4" direction="column">
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
  variant = "section",
}: {
  className?: string;
  hideDescription?: boolean;
  variant?: PricingVariant;
}) => {
  const [billing, setBilling] = useState<Billing>("annual");
  const isPage = variant === "page";

  return (
    <Box
      className={cn(
        "relative",
        isPage ? "py-12 md:pt-24 md:pb-16" : "md:pt-12",
        className,
      )}
    >
      {/* Decorative colour blurs intentionally removed for the warmer background. */}
      <Container className="max-w-332">
        <Box
          className={cn(
            "mt-10 flex flex-col gap-6 pb-6 md:flex-row md:items-end md:justify-between md:gap-16",
            isPage && "mt-0 max-w-3xl pb-0 md:mt-0",
          )}
        >
          <Box
            className={cn(isPage && "w-full text-center md:text-left")}
            data-landing-reveal
          >
            <Text
              as={isPage ? "h1" : "h2"}
              className={cn("mt-6 max-w-3xl pb-2 text-4xl md:text-5xl", {
                "mx-auto mt-0 max-w-2xl text-[2rem] leading-[1.05] font-medium tracking-[-0.04em] md:mx-0 md:text-[3.5rem]":
                  isPage,
              })}
            >
              Start with one team. Scale when the work does.
            </Text>
            {!hideDescription && isPage ? (
              <Text className="text-text-muted mt-4 hidden max-w-2xl text-base leading-6 font-normal md:block">
                No card and no trial clock. Run a real project on the free plan,
                then add teams, goals, integrations, and AI agent capacity as
                your organization grows.
              </Text>
            ) : null}
            {isPage ? (
              <Box className="mt-5 md:mt-7">
                <BillingSelector
                  billing={billing}
                  onChange={setBilling}
                  pricingPageStyle
                  showSavings
                />
              </Box>
            ) : null}
          </Box>
          {!hideDescription && !isPage ? (
            <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
              <Text className="w-full max-w-xl opacity-70 md:mt-4">
                No card and no trial clock. Run a real project on the free plan,
                then add teams, goals, integrations, and AI agent capacity as
                your organization grows.
              </Text>
            </Box>
          ) : null}
        </Box>

        {!isPage ? (
          <Flex className="mb-10" direction="column">
            <Box data-landing-reveal>
              <BillingSelector
                billing={billing}
                onChange={setBilling}
                showBillingSuffix
              />
            </Box>
          </Flex>
        ) : null}
        <Box
          className={cn("grid grid-cols-1 gap-5 md:grid-cols-4", {
            "mt-10 gap-y-12 pt-2 md:mt-20 md:gap-6": isPage,
          })}
          data-landing-reveal
          style={{ transitionDelay: "70ms" }}
        >
          {packages.map((pkg) => (
            <Package
              billing={billing}
              cta={pkg.cta}
              featureGroups={pkg.featureGroups}
              features={pkg.features}
              href={pkg.href}
              key={pkg.name}
              name={pkg.name}
              overview={pkg.overview}
              price={pkg.price}
              recommended={pkg.recommended}
              variant={variant}
            />
          ))}
        </Box>
      </Container>
    </Box>
  );
};
