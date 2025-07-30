import { Box, Text, Input, TextArea, Button } from "ui";

export const ContactForm = () => {
  return (
    <Box className="flex flex-col gap-6">
      <Box>
        <Text as="h2" className="text-4xl font-bold">
          SEND US A MESSAGE
        </Text>
        <Text className="mt-2 text-lg">
          We are always happy to hear from you.
        </Text>
      </Box>

      <Box className="flex flex-col gap-4">
        <Box>
          <Text className="k mb-2 font-medium">Full name*</Text>
          <Input
            className="w-full rounded-lg border border-gray-300 px-4 py-3"
            placeholder="Enter your full name"
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
            placeholder="Enter your message"
            rows={3}
          />
        </Box>
        <Button size="lg">Send message</Button>
      </Box>
    </Box>
  );
};
