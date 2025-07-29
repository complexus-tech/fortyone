import { BlurImage, Box, Button, Flex, Input, Text } from "ui";
import Image from "next/image";
import { SearchIcon } from "icons";
import { Container } from "@/components/ui";
import educationSvg from "../../../../public/images/home/education.svg";
import girlsRightsSvg from "../../../../public/images/home/girl.svg";

export const Hero = () => {
  return (
    <Container>
      <Box className="relative">
        <BlurImage
          className="aspect-[16/10]"
          quality={100}
          src="/images/home/hero.webp"
        />
        <Box className="absolute inset-0 grid grid-cols-2 bg-black/10">
          <Flex className="px-10 py-16" direction="column" justify="end">
            <Text as="h1" className="mb-4 text-5xl font-semibold text-white">
              Exisiting Partner
            </Text>
            <Text className="mb-6 text-lg text-white">
              Then search for the organisation to give towards
            </Text>
            <Input
              className="rounded-xl bg-white"
              leftIcon={<SearchIcon />}
              placeholder="Search organisation by name, location or cause"
              size="lg"
            />
            <Button className="ml-auto mt-4" color="secondary" size="lg">
              Find organisation
            </Button>
          </Flex>
          <Flex
            className="bg-secondary/50"
            direction="column"
            justify="between"
          >
            <Box className="p-10">
              <Text as="h1" className="mb-8 text-7xl font-semibold text-white">
                Be part of the change. <br /> Give today.
              </Text>
              <Text className="max-w-xs text-lg text-white">
                Your donation can help transform lives and bring hope to
                communities.
              </Text>
            </Box>
            <Box className="grid grid-cols-3">
              <Flex className="h-44 bg-secondary" direction="column">
                <Image
                  alt="Education"
                  className="h-24 w-auto invert"
                  src={educationSvg}
                />
                <Text className="text-center text-2xl text-white">
                  Education
                </Text>
              </Flex>
              <Flex
                className="h-44 bg-primary"
                direction="column"
                gap={4}
                justify="center"
              >
                <Image
                  alt="Education"
                  className="h-24 w-auto invert"
                  src={girlsRightsSvg}
                />
                <Text className="text-center text-2xl text-white">
                  Girls rights
                </Text>
              </Flex>
              <Flex
                className="h-44 bg-secondary"
                direction="column"
                justify="center"
              >
                <Text className="text-center text-2xl text-white">
                  Livelihoods & Economic Inclusion
                </Text>
              </Flex>
              <Flex
                className="h-44 bg-primary"
                direction="column"
                justify="center"
              >
                <Text className="text-center text-2xl text-white">
                  Food Security
                </Text>
              </Flex>
              <Flex
                className="h-44 bg-secondary"
                direction="column"
                justify="center"
              >
                <Text className="text-center text-2xl text-white">
                  Climate Change
                </Text>
              </Flex>
              <Flex
                className="h-44 bg-primary"
                direction="column"
                justify="center"
              >
                <Text className="text-center text-2xl text-white">
                  Youth Empowerment
                </Text>
              </Flex>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Container>
  );
};
