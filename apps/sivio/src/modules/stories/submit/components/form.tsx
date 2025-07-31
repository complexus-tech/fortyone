import { Box, Text, Input, TextArea, Button } from "ui";
import { DownloadIcon } from "icons";
import { Container } from "@/components/ui";

export const Form = () => {
  return (
    <Container className="pb-16">
      <Box className="mx-auto max-w-3xl">
        <Text as="h2" className="mb-8 text-3xl font-bold">
          SEND US A STORY
        </Text>

        <Box className="flex flex-col gap-6">
          <Box>
            <Text className="mb-2 font-medium">Full name of organisation*</Text>
            <Input
              className="w-full rounded-lg border border-gray-300 px-4 py-3"
              placeholder="Enter your organisation name"
              type="text"
            />
          </Box>
          <Box>
            <Text className="mb-2 font-medium">Email*</Text>
            <Input
              className="w-full rounded-lg border border-gray-300 px-4 py-3"
              placeholder="Enter your email"
              type="email"
            />
          </Box>
          <Box>
            <Text className="mb-2 font-medium">Message*</Text>
            <TextArea
              className="w-full rounded-lg border border-gray-300 px-4 py-3"
              placeholder="Share your story of impact..."
              rows={3}
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">
              Upload images or accompanying information
            </Text>
            <Box className="flex w-full items-center gap-3 rounded-lg bg-gray-100 px-6 py-3">
              <DownloadIcon />
              <Text className="text-gray-500">Click to upload files</Text>
            </Box>
          </Box>

          <Button size="lg">Send message</Button>
        </Box>
      </Box>
    </Container>
  );
};
