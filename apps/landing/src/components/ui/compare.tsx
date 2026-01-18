import { Flex, Text, Box, Button } from "ui";
import { CloseIcon, SuccessIcon } from "icons";
import { cn } from "lib";
import { plans, featureLabels } from "../../lib/plan-data";
import { Container } from "./container";

const FeatureCheck = ({ available }: { available: boolean }) => (
  <Box className="flex">
    {available ? (
      <SuccessIcon className="text-foreground" />
    ) : (
      <CloseIcon className="h-[1.15rem]" strokeWidth={2} />
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
    return <CloseIcon className="h-[1.15rem]" strokeWidth={2.3} />;
  }

  if (typeof value === "boolean") {
    return <FeatureCheck available={value} />;
  }

  return <Text>{value}</Text>;
};

export const ComparePlans = () => {
  return (
    <Container className="max-w-332 md:py-36">
      <Box className="hidden overflow-x-auto md:block">
        <Box>
          <Flex className="border-border border-b">
            <Box className="w-1/3 p-6" />
            {plans.map((plan) => (
              <Box
                className={cn("w-1/6 px-4 py-5", {
                  "border-border bg-surface-muted d dark:bg-surface-elevated rounded-t-2xl border border-b-0":
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
            <Flex className="border-border border-b">
              <Box className="w-1/3 px-4 py-3">
                <Text fontWeight="semibold">Limits</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-3", {
                    "border-border bg-surface-muted border-x":
                      plan.name === "Business",
                  })}
                  key={plan.name}
                />
              ))}
            </Flex>
            <Flex className="border-border border-b">
              <Box className="w-1/3 px-4 py-3">
                <Text>Members</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn(
                    "w-1/6 px-4 py-3",
                    plan.highlighted &&
                      "border-border bg-surface-muted border-x",
                  )}
                  key={`${plan.name}-members`}
                >
                  <Text>{plan.limits.members}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-border border-b">
              <Box className="w-1/3 px-4 py-3">
                <Text>File uploads</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-3", {
                    "border-border bg-surface-muted border-x": plan.highlighted,
                  })}
                  key={`${plan.name}-files`}
                >
                  <Text>{plan.limits.fileUploads}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-border border-b">
              <Box className="w-1/3 px-4 py-3">
                <Text>Stories</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-3", {
                    "border-border bg-surface-muted border-x": plan.highlighted,
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
            <Flex className="border-border border-b">
              <Box className="w-1/3 px-4 pt-8 pb-3">
                <Text fontWeight="semibold">Features</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-4 py-3", {
                    "border-border bg-surface-muted border-x":
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
                    className={!isLastItem ? "border-border border-b" : ""}
                    key={featureKey}
                  >
                    <Box className="w-1/3 px-4 py-3">
                      <Text>{featureLabel}</Text>
                    </Box>
                    {plans.map((plan) => {
                      const value =
                        plan.features[featureKey as keyof typeof plan.features];
                      const isHighlighted = plan.highlighted;
                      return (
                        <Box
                          className={cn("w-1/6 px-4 py-3", {
                            "border-border bg-surface-muted border-x":
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
          <Flex className="border-border border-t">
            <Box className="w-1/3" />
            <Box className="w-1/6 px-4 py-3">
              <Button
                align="center"
                color="tertiary"
                fullWidth
                href="https://cloud.fortyone.app/signup"
                variant="outline"
              >
                Start for free
              </Button>
            </Box>
            <Box className="w-1/6 px-4 py-3">
              <Button
                align="center"
                color="tertiary"
                fullWidth
                href="https://cloud.fortyone.app/signup"
                variant="outline"
              >
                Try Proffesional
              </Button>
            </Box>
            <Box className="border-border bg-surface-muted d dark:bg-surface-elevated w-1/6 rounded-b-2xl border border-t-0 px-4 py-3">
              <Button
                align="center"
                color="invert"
                fullWidth
                href="https://cloud.fortyone.app/signup"
                variant="outline"
              >
                Try Business
              </Button>
            </Box>
            <Box className="w-1/6 px-4 py-3">
              <Button
                align="center"
                color="tertiary"
                fullWidth
                href="mailto:info@complexus.app"
                variant="outline"
              >
                Contact sales
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Container>
  );
};
