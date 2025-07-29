import { Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";

const IndividualIcon = () => (
  <Box className="mt-4 flex items-center justify-center">
    <Box className="flex h-20 w-20 items-center justify-center rounded-full border-4 border-white">
      <svg fill="none" height="48" width="48">
        <circle cx="24" cy="24" r="22" stroke="white" strokeWidth="4" />
        <rect fill="white" height="8" rx="4" width="20" x="14" y="30" />
        <circle cx="24" cy="20" fill="white" r="8" />
      </svg>
    </Box>
  </Box>
);

const CorporateIcon = () => (
  <Box className="mt-4 flex items-center justify-center">
    <Box className="flex h-20 w-20 items-center justify-center rounded-full border-4 border-white">
      <svg fill="none" height="48" width="48">
        <rect fill="white" height="20" width="10" x="8" y="20" />
        <rect fill="white" height="28" width="10" x="20" y="12" />
        <rect fill="white" height="14" width="8" x="32" y="26" />
        <rect fill="white" height="4" width="24" x="12" y="36" />
      </svg>
    </Box>
  </Box>
);

const FeatureCard = ({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) => (
  <Flex align="center" className="flex-1 px-4 py-8" direction="column">
    <Text
      className="mb-2 text-4xl font-semibold text-white"
      style={{ fontFamily: "cursive" }}
    >
      {label}
    </Text>
    {children}
  </Flex>
);

export const WhatWeDo = () => (
  <Container className="py-28">
    <Flex align="start" className="flex-col gap-8 md:flex-row">
      <Box className="w-full md:w-1/3">
        <Box className="mb-4 h-2 w-36 bg-secondary" />
        <Text as="h2" className="text-6xl font-black leading-tight text-black">
          WHAT
          <br />
          WE DO
        </Text>
      </Box>
      <Box className="flex w-full flex-col gap-8 md:w-2/3">
        <Box className="pl-12">
          <Text as="h3" className="mb-2 text-5xl font-black text-black">
            We Provide
          </Text>
          <Text className="text-2xl leading-snug text-gray">
            an easy, credible and safe way to give to
            <br />
            causes across the continent for
          </Text>
        </Box>
        <Flex className="mt-6 flex-col items-stretch justify-center overflow-hidden bg-secondary sm:flex-row">
          <FeatureCard label="Individuals">
            <IndividualIcon />
          </FeatureCard>
          <FeatureCard label="Corporates">
            <CorporateIcon />
          </FeatureCard>
        </Flex>
      </Box>
    </Flex>
  </Container>
);
