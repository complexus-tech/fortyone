"use client";
import { Flex, Text, Box, Button } from "ui";
import { ErrorIcon, SuccessIcon } from "icons";
import { cn } from "lib";
import { motion } from "framer-motion";
import { useState } from "react";
import { plans, featureLabels } from "./plan-data";

const FeatureCheck = ({ available }: { available: boolean }) => (
  <Box className="flex">
    {available ? (
      <SuccessIcon className="text-primary dark:text-primary" />
    ) : (
      <ErrorIcon className="dark:text-white" />
    )}
  </Box>
);

// Helper to render a feature value
const FeatureValue = ({
  value,
}: {
  value: boolean | string | undefined | null;
}) => {
  if (value === undefined || value === null) {
    return <ErrorIcon className="dark:text-white" />;
  }

  if (typeof value === "boolean") {
    return <FeatureCheck available={value} />;
  }

  return <Text>{value}</Text>;
};

type Billing = "annual" | "monthly";
export const Plans = () => {
  const [billing, setBilling] = useState<Billing>("annual");
  return (
    <Box className="hidden overflow-x-auto md:block">
      <Box>
        <Flex className="border-b border-gray-100 dark:border-dark-100">
          <Box className="flex w-1/5 items-end px-3 py-6">
            <motion.div
              initial={{ y: 20, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.3,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Box className="flex w-full gap-1 rounded-[0.7rem] border border-gray-100 bg-gray-50 p-1 dark:border-dark-100 dark:bg-dark-300">
                {["annual", "monthly"].map((option) => (
                  <Button
                    className={cn("px-4 capitalize", {
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
                    {option}
                  </Button>
                ))}
              </Box>
            </motion.div>
          </Box>
          <Box className="w-1/5 px-4 py-6">
            <Text className="mb-2 text-2xl">Hobby</Text>
            <Text className="mb-2 text-3xl font-semibold">
              $0
              <Text as="span" color="muted" fontSize="lg">
                /mo
              </Text>
            </Text>
            <Button
              align="center"
              color="tertiary"
              disabled
              fullWidth
              href="/signup"
            >
              Current plan
            </Button>
          </Box>
          <Box className="w-1/5 px-4 py-6">
            <Text className="mb-2 text-2xl">Pro</Text>
            <Text className="mb-2 text-3xl font-semibold">
              $10
              <Text as="span" color="muted" fontSize="lg">
                /mo
              </Text>
            </Text>
            <Button align="center" color="tertiary" fullWidth href="/signup">
              Upgrade
            </Button>
          </Box>
          <Box className="w-1/5 rounded-t-2xl border border-b-0 border-gray-100 bg-gray-50 px-4 py-6 dark:border-dark-100 dark:bg-dark-300">
            <Text className="mb-2 text-2xl">Business</Text>
            <Text className="mb-2 text-3xl font-semibold">
              $20
              <Text as="span" color="muted" fontSize="lg">
                /mo
              </Text>
            </Text>
            <Button align="center" fullWidth href="/signup">
              Upgrade
            </Button>
          </Box>
          <Box className="w-1/5 px-4 py-6">
            <Text className="mb-2 text-2xl">Enterprise</Text>
            <Text className="mb-2 text-3xl font-semibold">Custom</Text>
            <Button align="center" color="tertiary" fullWidth href="/signup">
              Contact sales
            </Button>
          </Box>
        </Flex>

        {/* Limits section */}
        <Box>
          <Flex className="border-b border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300/50">
            <Box className="w-1/5 px-4 py-4">
              <Text fontWeight="semibold">Limits</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.name === "Business",
                })}
                key={plan.name}
              />
            ))}
          </Flex>
          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/5 px-4 py-4">
              <Text className="opacity-80">Members</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn(
                  "w-1/5 px-4 py-4",
                  plan.highlighted &&
                    "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300",
                )}
                key={`${plan.name}-members`}
              >
                <Text>{plan.limits.members}</Text>
              </Box>
            ))}
          </Flex>

          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/5 px-4 py-4">
              <Text className="opacity-80">File uploads</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.highlighted,
                })}
                key={`${plan.name}-files`}
              >
                <Text>{plan.limits.fileUploads}</Text>
              </Box>
            ))}
          </Flex>

          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/5 px-4 py-4">
              <Text className="opacity-80">Stories</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.highlighted,
                })}
                key={`${plan.name}stories`}
              >
                <Text>{plan.limits.issues}</Text>
              </Box>
            ))}
          </Flex>
        </Box>

        {/* Features section */}
        <Box>
          <Flex className="border-b border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300/50">
            <Box className="w-1/5 px-4 py-4">
              <Text fontWeight="semibold">Features</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.name === "Business",
                })}
                key={plan.name}
              />
            ))}
          </Flex>

          {/* Generate all possible feature rows */}
          {Object.entries(featureLabels).map(
            ([featureKey, featureLabel], index, array) => {
              // Skip 'teams' feature since it's already in limits section
              if (featureKey === "teams") return null;

              const isLastItem = index === array.length - 1;

              return (
                <Flex
                  className={
                    !isLastItem
                      ? "border-b border-gray-100 dark:border-dark-100"
                      : ""
                  }
                  key={featureKey}
                >
                  <Box className="w-1/5 px-4 py-4">
                    <Text className="opacity-80">{featureLabel}</Text>
                  </Box>
                  {plans.map((plan) => {
                    const value =
                      plan.features[featureKey as keyof typeof plan.features];
                    const isHighlighted = plan.highlighted;
                    return (
                      <Box
                        className={cn("w-1/5 px-4 py-4", {
                          "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                            isHighlighted,
                        })}
                        key={`${plan.name}-${featureKey}`}
                      >
                        <FeatureValue value={value} />
                      </Box>
                    );
                  })}
                </Flex>
              );
            },
          )}
        </Box>
      </Box>
    </Box>
  );
};
