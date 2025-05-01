"use client";

import { Box, Text, Flex, Button, BlurImage } from "ui";
import { motion } from "framer-motion";
import { cn } from "lib";
import { Container } from "@/components/ui";

const features = [
  {
    id: 1,
    title: "Create Stories Effortlessly",
    description:
      "Start by creating user stories that capture your requirements. Add acceptance criteria, and priority levels to ensure clarity.",
    image: "/objective.webp",
  },
  {
    id: 2,
    title: "Organize & Plan Visually",
    description:
      "Organize stories in objectives and sprints. Use our planning tools to estimate effort and create a realistic project timeline.",
    image: "/objective.webp",
  },
  {
    id: 3,
    title: "Track Progress in Real-time",
    description:
      "Monitor story progress with our intuitive dashboard. Get real-time updates on story status, team velocity, and objective health.",
    image: "/objective.webp",
  },
  {
    id: 4,
    title: "Collaborate & Deliver Together",
    description:
      "Work together seamlessly with built-in collaboration tools. Share updates, provide feedback, and celebrate wins as a team.",
    image: "/objective.webp",
  },
];

export const HowItWorks = () => {
  return (
    <Box className="relative">
      <Container className="py-20">
        <Flex align="center" className="mb-16" direction="column">
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.2,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text as="h2" className="text-5xl font-semibold md:text-7xl">
              How <span className="text-stroke-white">Stories</span> Work
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
              className="mt-6 max-w-xl text-center opacity-80 md:mt-8 md:text-xl"
              fontWeight="normal"
            >
              Our story management system is designed with powerful features to
              meet your team&apos;s needs
            </Text>
          </motion.div>
        </Flex>

        <Box className="mx-auto grid max-w-6xl grid-cols-1 border-t border-dark-100 md:grid-cols-2">
          {features.map((feature, idx) => (
            <motion.div
              initial={{ y: 30, opacity: 0 }}
              key={feature.id}
              transition={{
                duration: 0.8,
                delay: 0.2 + feature.id * 0.1,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Box
                className={cn(
                  "flex h-full flex-col border-b border-dark-200 p-12",
                  {
                    "border-r pl-0": idx % 2 === 0,
                    "pr-0": idx % 2 === 1,
                  },
                )}
              >
                <Text as="h3" className="mb-3 text-3xl font-semibold">
                  {feature.title}
                </Text>

                <Text className="mb-6 opacity-80">{feature.description}</Text>

                <Box className="relative">
                  <BlurImage
                    alt={feature.title}
                    className="aspect-[5/3.5]"
                    imageClassName="object-cover"
                    quality={100}
                    src={feature.image}
                  />
                  <Box className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-transparent via-black/90 to-black" />
                </Box>
              </Box>
            </motion.div>
          ))}
        </Box>

        <Flex className="mt-16" justify="center">
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.8,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Button
              className="px-6 font-semibold"
              href="/signup"
              rounded="full"
              size="lg"
            >
              Get Started Free
            </Button>
          </motion.div>
        </Flex>
      </Container>
    </Box>
  );
};
