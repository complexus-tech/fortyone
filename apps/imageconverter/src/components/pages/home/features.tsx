"use client";
import { Box, Text } from "ui";
import { motion } from "framer-motion";
import Image, { StaticImageData } from "next/image";
import { Container, Blur } from "@/components/ui";
import apiImg from "../../../../public/images/api.png";
import multipleFormatsImg from "../../../../public/images/images.png";
import bulkConversionImg from "../../../../public/images/bulk.png";
import secureProcessingImg from "../../../../public/images/security.png";

const Intro = () => (
  <Box className="relative">
    <Box as="section" className="my-12 text-center md:mt-16">
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{
          duration: 1,
          delay: 0,
        }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          as="h3"
          className="mx-auto pb-2 font-satoshi text-5xl font-extrabold md:text-6xl md:leading-[1.1]"
        >
          <Text as="span" color="gradient">
            Elevate
          </Text>{" "}
          your image <br /> conversion workflow
        </Text>
      </motion.div>
      <motion.div
        initial={{ y: 20, opacity: 0 }}
        transition={{
          duration: 1,
          delay: 0.3,
        }}
        viewport={{ once: true, amount: 0.5 }}
        whileInView={{ y: 0, opacity: 1 }}
      >
        <Text
          className="mx-auto mt-6 max-w-xl"
          fontSize="xl"
          fontWeight="normal"
        >
          Discover why imageconverter is the preferred image conversion platform
          for designers and developers
        </Text>
      </motion.div>
    </Box>
    <Blur className="absolute left-1/2 right-1/2 top-0 z-[4] h-[230vh] w-[60vw] -translate-x-1/2 bg-warning/10 md:h-[60vw] dark:bg-warning/[0.05]" />
  </Box>
);

export const Features = () => {
  const features: {
    heading: string;
    description: string;
    image: StaticImageData;
  }[] = [
    {
      heading: "Multiple Formats",
      image: multipleFormatsImg,
      description:
        "Convert between a wide range of image formats including JPG, PNG, WebP, and more.",
    },
    {
      heading: "API Access",
      image: apiImg,
      description:
        "Integrate our powerful conversion capabilities directly into your applications with our robust API.",
    },
    {
      heading: "Secure Processing",
      image: secureProcessingImg,
      description:
        "Your files are encrypted and deleted after conversion for maximum privacy.",
    },
    {
      heading: "Bulk Conversion",
      image: bulkConversionImg,
      description: "Convert multiple images at once to save time and effort.",
    },
  ];

  return (
    <Container as="section" className="mb-20 md:mb-40">
      <Intro />
      <Box className="mx-auto grid max-w-3xl grid-cols-1 gap-10 md:grid-cols-2 md:gap-x-20 md:gap-y-16">
        {features.map(({ heading, description, image }) => (
          <Box
            className="border-t border-gray-200/30 pt-8 text-center md:pt-10 dark:border-dark-200"
            key={heading}
          >
            <Image
              src={image}
              className="mx-auto mb-6 h-16 w-auto"
              alt={heading}
            />
            <Text
              fontSize="lg"
              className="font-satoshi"
              fontWeight="bold"
              transform="uppercase"
            >
              {heading}
            </Text>
            <Text className="mt-4 opacity-90" fontSize="lg" fontWeight="normal">
              {description}
            </Text>
          </Box>
        ))}
      </Box>
    </Container>
  );
};
