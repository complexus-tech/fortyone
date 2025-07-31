import { Box, Text, Button } from "ui";
import { DownloadIcon } from "icons";
import { Container } from "@/components/ui";

export const Content = () => {
  return (
    <Container className="py-16">
      <Box className="mx-auto max-w-3xl">
        <Text as="h2" className="mb-6 text-4xl font-bold text-black">
          Share Your Story of Impact
        </Text>

        <Text className="mb-8 text-lg text-black">
          We would love to hear how your work is making a difference. If your
          organisation has a story of impact worth sharing, you can submit it in
          one of three easy ways:
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
                with your story and any supporting materials.
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
          Your experiences inspire action. Let&apos;s amplify them together.
        </Text>

        <Button
          color="secondary"
          rightIcon={<DownloadIcon className="text-white" />}
          size="lg"
        >
          Download Template
        </Button>
      </Box>
    </Container>
  );
};
