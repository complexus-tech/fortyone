import { Flex, Text, Box, Button } from "ui";
import Image from "next/image";
import { Container, Blur } from "@/components/ui";

export const Integrations = () => {
  return (
    <Box className="relative bg-black py-32">
      <Image
        alt="Slack logo"
        className="absolute left-16 top-16 rotate-6"
        height={80}
        src="/integrations/slack.svg"
        width={80}
      />
      <Image
        alt="Intercom logo"
        className="absolute left-80 top-12 -rotate-6"
        height={75}
        src="/integrations/intercom.svg"
        width={75}
      />

      <Image
        alt="Notion logo"
        className="absolute bottom-1/2 left-48 top-1/2 -translate-y-1/2 rotate-12"
        height={95}
        src="/integrations/notion.svg"
        width={95}
      />
      <Image
        alt="Figma logo"
        className="absolute bottom-16 left-16 rotate-6"
        height={80}
        src="/integrations/figma.svg"
        width={80}
      />
      <Image
        alt="Github logo"
        className="absolute bottom-16 left-80 rotate-6"
        height={80}
        src="/integrations/github.svg"
        width={80}
      />

      <Container className="relative">
        <Flex
          align="center"
          className="md:mt-18 mb-8 text-center"
          direction="column"
        >
          <Text
            as="h1"
            className="mt-6 h-max max-w-4xl pb-2 text-7xl"
            color="gradient"
            fontWeight="semibold"
          >
            Sync up your favorite tools.
          </Text>
          <Text
            className="my-6 max-w-[600px] md:mt-6"
            color="muted"
            fontSize="xl"
            fontWeight="normal"
          >
            Complexus seamlessly integrates with popular tools like GitHub,
            Intercom, Figma, Notion, Slack, and Gitlab, empowering teams to
            achieve their objectives with greater efficiency.
          </Text>
          <Button color="tertiary" rounded="full" size="lg">
            View all integrations
          </Button>
        </Flex>
      </Container>
      <Image
        alt="Jira logo"
        className="absolute right-16 top-16 rotate-6"
        height={85}
        src="/integrations/jira.svg"
        width={85}
      />
      <Image
        alt="Drive logo"
        className="absolute right-80 top-12 -rotate-6"
        height={80}
        src="/integrations/drive.svg"
        width={80}
      />
      <Image
        alt="Gitlab logo"
        className="absolute bottom-1/2 right-48 top-1/2 -translate-y-1/2 rotate-12"
        height={95}
        src="/integrations/gitlab.svg"
        width={95}
      />
      <Image
        alt="Figma logo"
        className="absolute bottom-16 right-16 rotate-6"
        height={80}
        src="/integrations/teams.svg"
        width={80}
      />
      <Image
        alt="Zend logo"
        className="absolute bottom-16 right-80 rotate-6"
        height={80}
        src="/integrations/zend.svg"
        width={80}
      />
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[400px] w-[400px] -translate-x-1/2 -translate-y-1/2 bg-primary/40 dark:bg-info/5" />
    </Box>
  );
};
