import { Flex, Text, Box, Badge } from "ui";
import { ErrorIcon, SuccessIcon } from "icons";
import { cn } from "lib";
import { Container } from "./container";

type Plan = {
  name: string;
  highlighted?: boolean;
  limits: {
    members: string;
    fileUploads: string;
    issues: string;
    teams: string;
  };
  features: {
    issuesAndProjects: boolean;
    customerRequests: boolean;
    integrations: boolean;
    apiAccess: boolean;
    importExport: boolean;
    triage: boolean;
    issueSync: boolean;
    supportIntegrations: boolean;
  };
};

const plans: Plan[] = [
  {
    name: "Free",
    limits: {
      members: "Unlimited",
      fileUploads: "10MB",
      issues: "250",
      teams: "2",
    },
    features: {
      issuesAndProjects: true,
      customerRequests: true,
      integrations: true,
      apiAccess: true,
      importExport: true,
      triage: true,
      issueSync: true,
      supportIntegrations: false,
    },
  },
  {
    name: "Basic",
    limits: {
      members: "Unlimited",
      fileUploads: "Unlimited",
      issues: "Unlimited",
      teams: "5",
    },
    features: {
      issuesAndProjects: true,
      customerRequests: true,
      integrations: true,
      apiAccess: true,
      importExport: true,
      triage: true,
      issueSync: true,
      supportIntegrations: false,
    },
  },
  {
    name: "Business",
    highlighted: true,
    limits: {
      members: "Unlimited",
      fileUploads: "Unlimited",
      issues: "Unlimited",
      teams: "Unlimited",
    },
    features: {
      issuesAndProjects: true,
      customerRequests: true,
      integrations: true,
      apiAccess: true,
      importExport: true,
      triage: true,
      issueSync: true,
      supportIntegrations: true,
    },
  },
  {
    name: "Enterprise",
    limits: {
      members: "Unlimited",
      fileUploads: "Unlimited",
      issues: "Unlimited",
      teams: "Unlimited",
    },
    features: {
      issuesAndProjects: true,
      customerRequests: true,
      integrations: true,
      apiAccess: true,
      importExport: true,
      triage: true,
      issueSync: true,
      supportIntegrations: true,
    },
  },
];

const FeatureCheck = ({ available }: { available: boolean }) => (
  <Box className="flex">
    {available ? (
      <SuccessIcon className="dark:text-primary" />
    ) : (
      <ErrorIcon className="dark:text-white" />
    )}
  </Box>
);

export const ComparePlans = () => {
  return (
    <Container className="py-16 md:py-24">
      <Flex className="mb-8" direction="column">
        <Text className="text-4xl font-semibold md:text-5xl">
          Compare Plans
        </Text>
        <Text className="mt-4 max-w-2xl text-lg opacity-80">
          Choose the plan that&apos;s right for your team
        </Text>
      </Flex>

      {/* Mobile view - cards */}
      <Box className="grid grid-cols-1 gap-6 md:hidden">
        {plans.map((plan) => (
          <Box
            className={cn(
              "rounded-xl border border-dark-100 bg-dark-300 p-6",
              plan.highlighted && "border-primary shadow-md shadow-primary/20",
            )}
            key={plan.name}
          >
            <Flex align="center" className="mb-4">
              <Text className="text-xl font-semibold">{plan.name}</Text>
              {plan.highlighted ? (
                <Badge className="ml-2">Recommended</Badge>
              ) : null}
            </Flex>

            <Text className="mb-3 mt-6 font-medium">Limits</Text>
            <Flex className="mb-6" direction="column" gap={3}>
              <Flex justify="between">
                <Text>Members</Text>
                <Text>{plan.limits.members}</Text>
              </Flex>
              <Flex justify="between">
                <Text>File uploads</Text>
                <Text>{plan.limits.fileUploads}</Text>
              </Flex>
              <Flex justify="between">
                <Text>Issues</Text>
                <Text>{plan.limits.issues}</Text>
              </Flex>
              <Flex justify="between">
                <Text>Teams</Text>
                <Text>{plan.limits.teams}</Text>
              </Flex>
            </Flex>

            <Text className="mb-3 mt-6 font-medium">Features</Text>
            <Flex direction="column" gap={3}>
              <Flex justify="between">
                <Text>Issues, projects, cycles, and initiatives</Text>
                <FeatureCheck available={plan.features.issuesAndProjects} />
              </Flex>
              <Flex justify="between">
                <Text>Customer requests</Text>
                <FeatureCheck available={plan.features.customerRequests} />
              </Flex>
              <Flex justify="between">
                <Text>Integrations</Text>
                <FeatureCheck available={plan.features.integrations} />
              </Flex>
              <Flex justify="between">
                <Text>API and webhook access</Text>
                <FeatureCheck available={plan.features.apiAccess} />
              </Flex>
              <Flex justify="between">
                <Text>Import and export</Text>
                <FeatureCheck available={plan.features.importExport} />
              </Flex>
              <Flex justify="between">
                <Text>Triage</Text>
                <FeatureCheck available={plan.features.triage} />
              </Flex>
              <Flex justify="between">
                <Text>Issue sync</Text>
                <FeatureCheck available={plan.features.issueSync} />
              </Flex>
              <Flex justify="between">
                <Text>Support integrations</Text>
                <FeatureCheck available={plan.features.supportIntegrations} />
              </Flex>
            </Flex>
          </Box>
        ))}
      </Box>

      {/* Desktop view - table */}
      <Box className="hidden overflow-x-auto md:block">
        <Box>
          <Flex className="border-b border-dark-100">
            <Box className="w-1/3 p-6" />
            {plans.map((plan) => (
              <Box
                className={cn("w-1/6 px-6 py-5", {
                  "rounded-t-2xl border border-b-0 border-dark-100 bg-dark-300":
                    plan.highlighted,
                })}
                key={plan.name}
              >
                <Text className="text-lg font-semibold">{plan.name}</Text>
              </Box>
            ))}
          </Flex>

          {/* Limits section */}
          <Box>
            <Box className="border-b border-dark-100">
              <Text className="px-6 py-4 font-medium">Limits</Text>
            </Box>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Members</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn(
                    "w-1/6 px-6 py-4",
                    plan.highlighted && "border-x border-dark-100 bg-dark-300",
                  )}
                  key={`${plan.name}-members`}
                >
                  <Text>{plan.limits.members}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>File uploads</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-files`}
                >
                  <Text>{plan.limits.fileUploads}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Issues</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-issues`}
                >
                  <Text>{plan.limits.issues}</Text>
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Teams</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-teams`}
                >
                  <Text>{plan.limits.teams}</Text>
                </Box>
              ))}
            </Flex>
          </Box>

          {/* Features section */}
          <Box>
            <Box className="border-b border-dark-100">
              <Text className="px-6 py-4 font-medium">Features</Text>
            </Box>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Issues, projects, cycles, and initiatives</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-projects`}
                >
                  <FeatureCheck available={plan.features.issuesAndProjects} />
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Customer requests</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-requests`}
                >
                  <FeatureCheck available={plan.features.customerRequests} />
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Integrations</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-integrations`}
                >
                  <FeatureCheck available={plan.features.integrations} />
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>API and webhook access</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-api`}
                >
                  <FeatureCheck available={plan.features.apiAccess} />
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Import and export</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-import`}
                >
                  <FeatureCheck available={plan.features.importExport} />
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Triage</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-triage`}
                >
                  <FeatureCheck available={plan.features.triage} />
                </Box>
              ))}
            </Flex>

            <Flex className="border-b border-dark-100">
              <Box className="w-1/3 px-6 py-4">
                <Text>Issue sync</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "border-x border-dark-100 bg-dark-300": plan.highlighted,
                  })}
                  key={`${plan.name}-sync`}
                >
                  <FeatureCheck available={plan.features.issueSync} />
                </Box>
              ))}
            </Flex>

            <Flex>
              <Box className="w-1/3 p-6">
                <Text>Support integrations</Text>
              </Box>
              {plans.map((plan) => (
                <Box
                  className={cn("w-1/6 px-6 py-4", {
                    "rounded-b-2xl border border-t-0 border-dark-100 bg-dark-300":
                      plan.highlighted,
                  })}
                  key={`${plan.name}-support`}
                >
                  <FeatureCheck available={plan.features.supportIntegrations} />
                </Box>
              ))}
            </Flex>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
