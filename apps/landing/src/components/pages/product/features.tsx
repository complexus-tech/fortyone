import { Box, Flex, Text } from "ui";
import Image from "next/image";
import { Container, Blur } from "@/components/ui";
import { features } from "./fixtures";

export const Features = () => {
  return (
    <Container className="max-w-4xl">
      {features.map(
        ({ name, title, overview, icon, breakdown, image }, idx) => (
          <Box
            className="relative scroll-mt-12 border-t border-gray-200/5 pt-8 md:py-12"
            id={name.toLowerCase()}
            key={name}
          >
            <Flex
              align="center"
              className="relative -left-1 mb-2 opacity-15 md:mb-6"
              gap={6}
              justify="between"
            >
              <Text
                className="text-stroke-white text-[2.6rem] md:text-6xl"
                fontWeight="bold"
              >
                <span className="hidden md:inline">0{idx + 1}. </span>
                {name}
              </Text>
              {icon}
            </Flex>
            <Box className="col-span-3">
              <Text className="text-xl opacity-80 md:text-3xl">{title}</Text>
              <Text className="my-2 md:my-4" color="muted" fontSize="lg">
                {overview}
              </Text>
              <Box className="rounded-xl bg-dark-100 p-1 md:p-1.5">
                <Image
                  alt={name}
                  className="pointer-events-none rounded-lg"
                  placeholder="blur"
                  src={image}
                />
              </Box>
              <Box className="mt-6 px-0.5">
                {breakdown.map(({ title: subTitle, overview: subOverview }) => (
                  <Box className="" key={subTitle}>
                    <Text className="mb-2" fontSize="lg">
                      {subTitle}
                    </Text>
                    <Text className="mb-5" color="muted">
                      {subOverview}
                    </Text>
                  </Box>
                ))}
              </Box>
            </Box>
            <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[650px] w-[650px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
          </Box>
        ),
      )}
      <Text
        className="mb-16 mt-12 text-xl leading-snug opacity-80 md:mb-28 md:mt-0 md:text-2xl"
        fontWeight="normal"
      >
        Complexus is tailored to equip you with comprehensive tools for
        effective project management. Whether you&rsquo;re tracking tasks,
        visualizing project roadmaps, or generating reports, Complexus has
        everything you need to succeed.
      </Text>
    </Container>
  );
};
