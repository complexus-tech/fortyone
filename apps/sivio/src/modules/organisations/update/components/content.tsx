import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Content = () => {
  return (
    <Container className="pt-16">
      <Box className="mx-auto max-w-3xl">
        <Text as="h2" className="mb-6 text-4xl font-bold text-black">
          Need to update your AfricaGiving profile?
        </Text>

        <Text className="mb-8 text-lg text-black">
          You have a few convenient options:
        </Text>

        <Box className="mb-8 flex flex-col gap-4">
          <ul className="list-disc pl-8">
            <li>
              <Text className="text-lg">
                Email us at{" "}
                <a
                  className="text-primary"
                  href="mailto:africagiving@sivioinstitute.org"
                >
                  africagiving@sivioinstitute.org
                </a>{" "}
                with the updated details.
              </Text>
            </li>
            <li>
              <Text className="text-lg">
                Download the story submission template, fill it out, and send it
                back to us.
              </Text>
            </li>
            <li>
              <Text className="text-lg">
                Use the online portal below to share your story directly.
              </Text>
            </li>
          </ul>
        </Box>

        <Text className="mb-6 text-lg text-black">
          We are here to keep your profile accurate and up to date!
        </Text>
      </Box>
    </Container>
  );
};
