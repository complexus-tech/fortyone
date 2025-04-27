import { Flex, Text, Box, Button } from "ui";
import { ErrorIcon, SuccessIcon } from "icons";
import { cn } from "lib";
import { plans, featureLabels } from "../../lib/plan-data";
import { Container } from "./container";

const FeatureCheck = ({ available }: { available: boolean }) => (
  <Box className="flex">
    {available ? (
      <SuccessIcon className="dark:text-primary" />
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

export const ComparePlans = () => {
  return (
    <Container className="md:py-24">
      <Box className="hidden overflow-x-auto md:block">
        <Box>
          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/3 p-6" />
            {plans.map((plan) => (
              <Box
                className={cn("w-1/6 px-4 py-5", {
                  "rounded-t-2xl border border-b-0 border-dark-100 bg-dark-300":
                    plan.highlighted,
                })}
                key={plan.name}
              >
                <Text className="text-xl font-semibold">{plan.name}</Text>
              </Box>
            ))}
          </Flex>

          {/* Limits section */}
          <Box>
            <Flex className="border-b border-dark-100 bg-dark-300/50">
              <Box className="w-1/3 px-4 py-4">
                <Text fontWeight="semibold">Limits</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-4", {
                    "border-x border-dark-100 bg-dark-300":
                      plan.name === "Business",
                  })}
                  key={plan.name}
                />
              ))}
            </Flex>
            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-4 py-4">
                <Text>Members</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn(
                    "w-1/6 px-4 py-4",
                    plan.highlighted && "border-x border-dark-100 bg-dark-300",
                  )}
                  key={`${plan.name}-members`}
                >
                  <Text>{plan.limits.members}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-4 py-4">
                <Text>File uploads</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-files`}
                >
                  <Text>{plan.limits.fileUploads}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-4 py-4">
                <Text>Stories</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
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
            <Flex className="border-b border-dark-100 bg-dark-300/50">
              <Box className="w-1/3 px-4 py-4">
                <Text fontWeight="semibold">Features</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-4", {
                    "border-x border-dark-100 bg-dark-300":
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
                    className={!isLastItem ? "border-b border-dark-100" : ""}
                    key={featureKey}
                  >
                    <Box className="w-1/3 px-4 py-4">
                      <Text>{featureLabel}</Text>
                    </Box>
                    {plans.map((plan) => {
                      const value =
                        plan.features[featureKey as keyof typeof plan.features];
                      const isHighlighted = plan.highlighted;
                      return (
                        <Box
                          className={cn("w-1/6 px-4 py-4", {
                            "border-x border-dark-100 bg-dark-300":
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
          <Flex className="border-t border-dark-100">
            <Box className="w-1/3" />
            <Box className="w-1/6 px-4 py-4">
              <Button align="center" color="tertiary" fullWidth href="/signup">
                Start for free
              </Button>
            </Box>
            <Box className="w-1/6 px-4 py-4">
              <Button align="center" color="tertiary" fullWidth href="/signup">
                Upgrade now
              </Button>
            </Box>
            <Box className="w-1/6 rounded-b-2xl border border-t-0 border-dark-100 bg-dark-300 px-4 py-4">
              <Button align="center" fullWidth href="/signup">
                Upgrade now
              </Button>
            </Box>
            <Box className="w-1/6 px-4 py-4">
              <Button align="center" color="tertiary" fullWidth href="/signup">
                Contact sales
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Container>
  );
};
