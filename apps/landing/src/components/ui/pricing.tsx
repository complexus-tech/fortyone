import { Flex, Text, Box, Button, Badge } from "ui";
import { ArrowRightIcon, CheckIcon } from "icons";
import { Blur } from "./blur";
import { Container } from "./container";

export const Pricing = () => {
  const hobby = [
    "Up to 5 users",
    "Up to 5 guests",
    "Up to 2 teams",
    "Unlimited stories",
    "Unlimited objectives",
    "Import & export data",
    "Basic support",
    "2GB storage",
  ];

  const starter = [
    "Unlimited users",
    "Unlimited guests",
    "Unlimited teams",
    "Integrations & API",
    "Priority support",
    "Advanced reporting",
    "50GB storage",
  ];

  return (
    <Box className="relative mb-40">
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[700px] w-[700px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
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
          <Box className="rounded-3xl border-2 border-dark-100 bg-dark p-8 shadow-xl shadow-black">
            <Text className="mb-2 text-center text-2xl" fontWeight="medium">
              Hobby
            </Text>
            <Text
              className="text-center text-lg opacity-80"
              fontWeight="normal"
            >
              For small teams
            </Text>
            <Text className="mt-4 text-center text-4xl" fontWeight="medium">
              $0
              <Text as="span" className="text-lg opacity-80">
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
              Start for free
            </Button>
            <Text className="mt-6" color="muted">
              Start your hobby project with our free plan. No credit card
              required.
            </Text>
            <Flex className="mt-4" direction="column" gap={4}>
              {hobby.map((feature) => (
                <Flex align="center" gap={2} key={feature}>
                  <CheckIcon className="h-5 w-auto text-primary" />
                  <Text fontWeight="medium">{feature}</Text>
                </Flex>
              ))}
            </Flex>
          </Box>

          <Box className="rounded-3xl border-2 border-primary bg-dark p-8 shadow-2xl shadow-primary/20">
            <Text
              className="mb-2 flex items-center justify-center gap-1.5 text-2xl"
              fontWeight="medium"
            >
              Professional <Badge>Most Popular</Badge>
            </Text>
            <Text
              className="text-center text-lg opacity-80"
              fontWeight="normal"
            >
              For growing teams
            </Text>
            <Text className="mt-4 text-center text-4xl" fontWeight="medium">
              $9
              <Text as="span" className="text-lg opacity-80">
                user/mo
              </Text>
            </Text>
            <Button
              className="mt-6 w-full justify-between"
              color="primary"
              rightIcon={<ArrowRightIcon className="h-4 w-auto" />}
              rounded="full"
              size="lg"
            >
              Upgrage now
            </Button>

            <Text className="mt-6" color="muted">
              Everything in Hobby, plus more features for growing teams.
            </Text>
            <Flex className="mt-4" direction="column" gap={4}>
              {starter.map((feature) => (
                <Flex align="center" gap={2} key={feature}>
                  <CheckIcon className="h-5 w-auto text-primary" />
                  <Text fontWeight="medium">{feature}</Text>
                </Flex>
              ))}
            </Flex>
          </Box>

          <Box className="rounded-3xl border-2 border-dark-100 bg-dark p-8 shadow-xl shadow-black">
            <Text className="mb-2 text-center text-2xl" fontWeight="medium">
              Enteprise
            </Text>
            <Text
              className="text-center text-lg opacity-80"
              fontWeight="normal"
            >
              For large teams
            </Text>
            <Text className="mt-4 text-center text-4xl" fontWeight="medium">
              $15
              <Text as="span" className="text-lg opacity-80">
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
              Contact sales
            </Button>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
