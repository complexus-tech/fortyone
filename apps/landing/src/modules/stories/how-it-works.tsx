"use client";

import { Box, Text } from "ui";
import Image from "next/image";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

const steps = [
  {
    id: 1,
    name: "Create Stories",
    description:
      "Start by creating user stories that capture your requirements. Add acceptance criteria, story points, and priority levels to ensure clarity.",
    image: "/objective.webp",
  },
  {
    id: 2,
    name: "Organize & Plan",
    description:
      "Organize stories into epics and sprints. Use our planning tools to estimate effort and create a realistic project timeline.",
    image: "/objective.webp",
  },
  {
    id: 3,
    name: "Track Progress",
    description:
      "Monitor story progress with our intuitive dashboard. Get real-time updates on story status, team velocity, and project health.",
    image: "/objective.webp",
  },
  {
    id: 4,
    name: "Collaborate & Deliver",
    description:
      "Work together seamlessly with built-in collaboration tools. Share updates, provide feedback, and celebrate wins as a team.",
    image: "/objective.webp",
  },
];

export const HowItWorks = () => {
  return (
    <Box className="relative">
      <Container className="relative max-w-4xl pb-16 pt-36 md:pt-40">
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
            as="h2"
            className="text-5xl font-semibold md:max-w-5xl md:text-7xl"
          >
            How Stories Work in{" "}
            <span className="text-stroke-white">Complexus</span>
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
            className="mt-6 max-w-xl opacity-80 md:mt-0"
            fontSize="xl"
            fontWeight="normal"
          >
            Our story management system is designed to be intuitive and
            efficient. Here&apos;s how it works:
          </Text>
        </motion.div>
        <Box className="mx-auto mt-16 grid grid-cols-1 gap-6 md:grid-cols-2">
          {steps.map((step) => (
            <motion.div
              initial={{ y: 20, opacity: 0 }}
              key={step.id}
              transition={{
                duration: 1,
                delay: 0.3 + step.id * 0.1,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Box className="relative flex min-h-[400px] flex-col justify-between overflow-hidden rounded-3xl border border-dark-50 bg-dark p-6 pb-8 md:h-[420px]">
                <Box className="relative h-48 w-full overflow-hidden rounded-lg">
                  <Image
                    alt={step.name}
                    className="object-cover"
                    fill
                    src={step.image}
                  />
                  <Box className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-transparent via-dark/90 to-dark" />
                </Box>
                <Box>
                  <Text as="h3" className="text-lg font-semibold">
                    {step.name}
                  </Text>
                  <Text className="mt-4 opacity-80">{step.description}</Text>
                </Box>
                <div className="pointer-events-none absolute inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-50" />
              </Box>
            </motion.div>
          ))}
        </Box>
      </Container>
    </Box>
  );
};
