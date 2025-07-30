import React from "react";
import { Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";

export const Apply = () => {
  return (
    <Container>
      <Flex align="center" className="bg-primary p-16" justify="between">
        <Text className="text-center text-7xl font-bold uppercase text-white">
          Apply Here
        </Text>
      </Flex>

      <Box className="mx-auto my-10 max-w-5xl space-y-8">
        <Text className="text-lg">
          As a registered non-profit in your country of operation, you are
          welcome to join the AfricaGiving network and enable your supporters to
          give financially towards achieving your vision and mission. Below is
          the application form. Consider providing as much information as you
          are able to increase your transparency with potential givers.
        </Text>
        <Text className="text-lg">
          Once the application form has been submitted, the SIVIO Institute team
          will review the application and send an MOU to finalise the agreement
          on using AfricaGiving.
        </Text>
        <Text className="text-lg">
          If you have any questions about the application process, please do not
          hesitate to reach out to the team using
          <a
            className="underline"
            href="mailto:africagiving@sivioinstitute.org"
          >
            africagiving@sivioinstitute.org
          </a>
        </Text>
      </Box>
    </Container>
  );
};
