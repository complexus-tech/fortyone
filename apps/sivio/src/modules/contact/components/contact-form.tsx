import { Box, Text, Input, TextArea, Button } from "ui";

export const ContactForm = () => {
  return (
    <Box className="flex flex-col gap-6">
      <Box>
        <Text as="h2" className="text-4xl font-bold text-black">
          SEND US A MESSAGE
        </Text>
        <Text className="text-gray-600 mt-2 text-lg">
          We are always happy to hear from you.
        </Text>
      </Box>

      <Box className="flex flex-col gap-4">
        <Box>
          <Text className="mb-2 text-sm font-medium text-black">
            Full name*
          </Text>
          <Input
            type="text"
            placeholder="Enter your full name"
            className="w-full rounded-lg border border-gray-300 px-4 py-3"
          />
        </Box>

        <Box>
          <Text className="mb-2 text-sm font-medium text-black">Email*</Text>
          <Input
            type="email"
            placeholder="Enter your email"
            className="w-full rounded-lg border border-gray-300 px-4 py-3"
          />
        </Box>

        <Box>
          <Text className="mb-2 text-sm font-medium text-black">Message*</Text>
          <TextArea
            placeholder="Enter your message"
            rows={5}
            className="w-full rounded-lg border border-gray-300 px-4 py-3"
          />
        </Box>

        <Button className="bg-green-600 hover:bg-green-700 mt-4 px-6 py-3 text-white">
          Send message
        </Button>
      </Box>
    </Box>
  );
};
