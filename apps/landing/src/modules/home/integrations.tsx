"use client";
import { Flex, Text, Box, Button } from "ui";
import Image from "next/image";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";

export const Integrations = () => {
  return (
    <Box className="from-surface-muted to-surface dark:from-background dark:via-background relative mt-20 bg-linear-to-b py-16 md:mb-28 md:py-40 dark:to-black">
      <Image
        alt="Slack logo"
        className="pointer-events-none absolute top-24 left-16 hidden rotate-6 md:block"
        height={55}
        src="/integrations/slack.svg"
        width={55}
      />
      <Image
        alt="Intercom logo"
        className="pointer-events-none absolute top-20 left-80 hidden -rotate-6 md:block"
        height={55}
        src="/integrations/intercom-icon.svg"
        width={55}
      />

      <Image
        alt="Notion logo"
        className="pointer-events-none absolute top-1/2 bottom-1/2 left-48 hidden -translate-y-1/2 rotate-12 invert md:block dark:invert-0"
        height={55}
        src="/integrations/notion.svg"
        width={55}
      />
      <Image
        alt="Figma logo"
        className="pointer-events-none absolute bottom-24 left-16 hidden rotate-6 md:block"
        height={55}
        src="/integrations/figma.svg"
        width={55}
      />
      <Image
        alt="Github logo"
        className="pointer-events-none absolute bottom-24 left-80 hidden rotate-6 invert md:block dark:invert-0"
        height={55}
        src="/integrations/github.svg"
        width={55}
      />

      <Container className="relative">
        <Flex
          align="center"
          className="mb-4 text-center md:mt-18"
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
              className="h-max max-w-2xl pb-2 text-4xl md:mt-6 md:text-5xl"
            >
              Sync up your whole stack.
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
            <Text className="mt-4 mb-10 max-w-[550px]" color="muted">
              GitHub merges, Slack threads, Figma handoffs, GitLab pipelines —
              they all happen in isolation, then get lost. FortyOne pulls them
              into a single, coherent view so your team stops chasing context
              across five tabs.
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
            <Button color="invert" href={SIGNUP_URL} rounded="lg" size="lg">
              Get started
            </Button>
          </motion.div>
        </Flex>
      </Container>
      <Image
        alt="Jira logo"
        className="pointer-events-none absolute top-24 right-16 hidden rotate-6 md:block"
        height={55}
        src="/integrations/jira.svg"
        width={55}
      />
      <Image
        alt="Drive logo"
        className="pointer-events-none absolute top-20 right-80 hidden -rotate-6 md:block"
        height={55}
        src="/integrations/drive.svg"
        width={55}
      />
      <Image
        alt="Gitlab logo"
        className="pointer-events-none absolute top-1/2 right-48 bottom-1/2 hidden -translate-y-1/2 rotate-12 md:block"
        height={55}
        src="/integrations/gitlab.svg"
        width={55}
      />
      <Image
        alt="Figma logo"
        className="pointer-events-none absolute right-16 bottom-24 hidden rotate-6 md:block"
        height={55}
        src="/integrations/teams.svg"
        width={55}
      />
      <Image
        alt="Zend logo"
        className="pointer-events-none absolute right-80 bottom-24 hidden rotate-6 invert md:block dark:invert-0"
        height={55}
        src="/integrations/zend.svg"
        width={55}
      />
    </Box>
  );
};
