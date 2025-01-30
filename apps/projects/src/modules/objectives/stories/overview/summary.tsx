import { CalendarIcon, HealthIcon, OKRIcon } from "icons";
import { Box, Text, Wrapper, ProgressBar } from "ui";

export const Summary = () => {
  return (
    <Box className="mt-3 grid grid-cols-3 gap-4">
      <Wrapper className="px-5">
        <Text
          className="mb-2 flex items-center gap-1.5 antialiased"
          fontSize="lg"
          fontWeight="semibold"
        >
          <HealthIcon />
          Progress
        </Text>
        <Text fontSize="lg">
          33%{" "}
          <Text as="span" color="muted" fontSize="md">
            completed
          </Text>
        </Text>
        <ProgressBar className="mt-2" progress={33} />
      </Wrapper>
      <Wrapper className="px-5">
        <Text
          className="mb-2 flex items-center gap-1.5 antialiased"
          fontSize="lg"
          fontWeight="semibold"
        >
          <OKRIcon />
          Key Result Progress
        </Text>
        <Text fontSize="lg">
          65%{" "}
          <Text as="span" color="muted" fontSize="md">
            completed
          </Text>
        </Text>
        <ProgressBar className="mt-2" progress={65} />
      </Wrapper>
      <Wrapper className="px-5">
        <Text
          className="mb-2 flex items-center gap-1 antialiased"
          fontSize="lg"
          fontWeight="semibold"
        >
          <CalendarIcon />
          Target date
        </Text>
        <Text fontSize="lg">
          Feb 10
          <Text as="span" color="muted" fontSize="md">
            , 2025
          </Text>
        </Text>
        <ProgressBar className="mt-2" progress={50} />
      </Wrapper>
    </Box>
  );
};
