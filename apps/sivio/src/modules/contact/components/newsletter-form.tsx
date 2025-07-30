import { Box, Text, Input, Button } from "ui";

export const NewsletterForm = () => {
  return (
    <Box className="flex flex-col gap-6">
      <Box>
        <Text as="h2" className="text-4xl font-bold">
          STAY UP TO DATE
        </Text>
        <Text className="mt-2 text-lg">
          Subscribe to our newsletter to stay up to date with current
          information!
        </Text>
      </Box>

      <Box className="flex flex-col gap-4">
        <Box>
          <Text className="mb-2 font-medium">First name*</Text>
          <Input
            className="w-full rounded-lg border border-gray-300 px-4 py-3"
            placeholder="Enter your first name"
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

        <Button size="lg">Join newsletter</Button>

        <Text className="text-sm">
          *NB* We will not spam your inbox, and you can also unsubscribe at any
          time.
        </Text>
      </Box>
    </Box>
  );
};
