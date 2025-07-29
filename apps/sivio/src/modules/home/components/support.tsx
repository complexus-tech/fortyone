import { Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";

type StatCardProps = {
  value: string;
  label: string;
};

const StatCard = ({ value, label }: StatCardProps) => (
  <Flex align="center" className="flex-1 text-center" direction="column">
    <Box className="mb-4 size-20 rounded-full bg-black" />
    <Text className="mb-1 text-4xl font-black">{value}</Text>
    <Text as="div" className="text-lg">
      {label}
    </Text>
  </Flex>
);

export const Support = () => (
  <Container className="py-28">
    <Flex align="start" className="flex-col gap-8 md:flex-row">
      <Box className="w-full md:w-1/3">
        <Box className="mb-4 h-2 w-36 bg-secondary" />
        <Text as="h2" className="text-6xl font-black leading-tight text-black">
          WHO
          <br />
          ARE WE?
        </Text>
      </Box>

      <Box className="flex w-full md:w-2/3">
        <Box className="flex max-w-2xl flex-col gap-4">
          <Text as="h3" className="text-5xl font-black">
            We support Africa
          </Text>
          <Text as="h4" className="text-3xl font-semibold text-gray">
            through Smart Giving
          </Text>
          <Text className="text-lg text-gray">
            AfricaGiving, an initiative of{" "}
            <span className="font-bold underline">SIVIO Institute</span>, is a
            fundraising platform built to connect individuals and corporates who
            are interested in supporting change across Africa.
          </Text>
          <Text className="text-lg text-gray">
            Whether you are an individual, a corporate donor, or a foundation,
            AfricaGiving helps you channel your generosity to trusted,
            grassroots-led initiatives that are building a better Africa every
            day.
          </Text>
        </Box>
      </Box>
    </Flex>
    <Flex className="mt-20 flex-col gap-8 sm:flex-row">
      <StatCard label="organisations" value="135+" />
      <StatCard label="African countries" value="17" />
      <StatCard label="transparency verified" value="100%" />
    </Flex>
  </Container>
);
