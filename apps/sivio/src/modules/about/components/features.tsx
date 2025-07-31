import { BlurImage, Box, Text } from "ui";
import Image from "next/image";
import { Container } from "@/components/ui";
import educationSvg from "../../../../public/images/home/education.svg";

export const Features = () => {
  return (
    <Box className="relative">
      <BlurImage
        className="aspect-[16/7]"
        imageClassName="object-cover"
        src="/images/about/2.jpg"
      />
      <Box className="absolute inset-0 flex items-center bg-secondary/90 backdrop-blur-sm">
        <Container className="grid max-w-4xl grid-cols-1 gap-8 md:grid-cols-2">
          <Box className="flex flex-col items-center gap-4">
            <Image
              alt="Education"
              className="h-24 w-auto invert"
              src={educationSvg}
            />
            <Text as="h3" className="text-center text-3xl font-bold text-white">
              A Trusted Giving Platform
            </Text>
            <Text className="text-center text-lg text-white">
              Every organisation listed is vetted to ensure accountability,
              transparency, and community driven solutions.
            </Text>
          </Box>

          <Box className="flex flex-col items-center gap-4">
            <Image
              alt="Education"
              className="h-24 w-auto invert"
              src={educationSvg}
            />
            <Text as="h3" className="text-center text-3xl font-bold text-white">
              Giving Made Simple
            </Text>
            <Text className="text-center text-lg text-white">
              Donate quickly and securely using your preferred method of
              transfer from anywhere in the world.
            </Text>
          </Box>

          <Box className="flex flex-col items-center gap-4">
            <Image
              alt="Education"
              className="h-24 w-auto invert"
              src={educationSvg}
            />
            <Text as="h3" className="text-center text-3xl font-bold text-white">
              You Choose the Cause
            </Text>
            <Text className="text-center text-lg text-white">
              Support work in education, health, entrepreneurship, climate
              justice, gender equality, and more.
            </Text>
          </Box>

          <Box className="flex flex-col items-center gap-4">
            <Image
              alt="Education"
              className="h-24 w-auto invert"
              src={educationSvg}
            />
            <Text as="h3" className="text-center text-3xl font-bold text-white">
              Homegrown Impact
            </Text>
            <Text className="text-center text-lg text-white">
              AfricaGiving prioritises African-led organisations that know their
              communities and deliver sustainable results.
            </Text>
          </Box>
        </Container>
      </Box>
    </Box>
  );
};
