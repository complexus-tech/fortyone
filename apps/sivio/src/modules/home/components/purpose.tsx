import { Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";

export const Purpose = () => (
  <Container className="py-36">
    <Flex align="start" className="flex-col gap-8 md:flex-row">
      <Box className="w-full md:w-1/3">
        <Box className="mb-4 h-2 w-36 bg-secondary" />
        <Text as="h2" className="text-6xl font-black leading-tight text-black">
          OUR
          <br />
          PURPOSE
        </Text>
      </Box>
      <Box className="flex w-full items-start md:w-2/3">
        <Text
          as="span"
          className="mr-6 mt-[-1.5rem] select-none text-[14rem] font-black leading-none text-black"
        >
          &ldquo;
        </Text>
        <Text className="max-w-3xl text-3xl font-medium text-gray md:text-4xl">
          To build a culture of giving rooted in African values of community,
          solidarity, and shared responsibility, fuelled by trust and
          transparency.
        </Text>
      </Box>
    </Flex>
  </Container>
);
