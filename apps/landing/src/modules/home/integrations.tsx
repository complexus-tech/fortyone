"use client";
import { Flex, Text, Box, Button } from "ui";
import Image from "next/image";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";

export const Integrations = () => {
  return (
    <Box className="relative bg-gradient-to-b from-white to-white py-16 dark:from-black dark:via-dark dark:to-black md:py-40">
      <Image
        alt="Slack logo"
        className="pointer-events-none absolute left-16 top-24 hidden rotate-6 md:block"
        height={80}
        src="/integrations/slack.svg"
        width={80}
      />
      <Image
        alt="Intercom logo"
        className="pointer-events-none absolute left-80 top-20 hidden -rotate-6 md:block"
        height={75}
        src="/integrations/intercom-icon.svg"
        width={75}
      />

      <Image
        alt="Notion logo"
        className="pointer-events-none absolute bottom-1/2 left-48 top-1/2 hidden -translate-y-1/2 rotate-12 invert dark:invert-0 md:block"
        height={95}
        src="/integrations/notion.svg"
        width={95}
      />
      <Image
        alt="Figma logo"
        className="pointer-events-none absolute bottom-24 left-16 hidden rotate-6 md:block"
        height={80}
        src="/integrations/figma.svg"
        width={80}
      />
      <Image
        alt="Github logo"
        className="pointer-events-none absolute bottom-24 left-80 hidden rotate-6 invert dark:invert-0 md:block"
        height={80}
        src="/integrations/github.svg"
        width={80}
      />

      <Container className="relative">
        <Flex
          align="center"
          className="md:mt-18 mb-4 text-center"
          direction="column"
        >
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
              className="h-max max-w-2xl pb-2 text-5xl font-semibold md:mt-6 md:text-6xl"
            >
              <Text as="span">Sync up</Text> your favorite{" "}
              <span className="text-stroke-white">tools</span>.
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
              className="my-10 max-w-[600px] font-normal"
              color="muted"
              fontSize="xl"
            >
              Connect your entire stack with GitHub, GitLab, Figma, Slack, and
              more. Keep development, design, and communication unified.
            </Text>
          </motion.div>
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0.6,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Button color="invert" href="/product" rounded="lg" size="lg">
              View all integrations
            </Button>
          </motion.div>
        </Flex>
      </Container>
      <Image
        alt="Jira logo"
        className="pointer-events-none absolute right-16 top-24 hidden rotate-6 md:block"
        height={85}
        src="/integrations/jira.svg"
        width={85}
      />
      <Image
        alt="Drive logo"
        className="pointer-events-none absolute right-80 top-20 hidden -rotate-6 md:block"
        height={80}
        src="/integrations/drive.svg"
        width={80}
      />
      <Image
        alt="Gitlab logo"
        className="pointer-events-none absolute bottom-1/2 right-48 top-1/2 hidden -translate-y-1/2 rotate-12 md:block"
        height={95}
        src="/integrations/gitlab.svg"
        width={95}
      />
      <Image
        alt="Figma logo"
        className="pointer-events-none absolute bottom-24 right-16 hidden rotate-6 md:block"
        height={80}
        src="/integrations/teams.svg"
        width={80}
      />
      <Image
        alt="Zend logo"
        className="pointer-events-none absolute bottom-24 right-80 hidden rotate-6 invert dark:invert-0 md:block"
        height={80}
        src="/integrations/zend.svg"
        width={80}
      />
    </Box>
  );
};
