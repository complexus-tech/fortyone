import { Box, Text, Input, Button } from "ui";

export const NewsletterForm = () => {
  return (
    <Box className="flex flex-col gap-6">
      <Box>
        <Text as="h2" className="text-4xl font-bold text-black">
          STAY UP TO DATE
        </Text>
        <Text className="text-gray-600 mt-2 text-lg">
          Subscribe to our newsletter to stay up to date with current
          information!
        </Text>
      </Box>

      <Box className="flex flex-col gap-4">
        <Box>
          <Text className="mb-2 text-sm font-medium text-black">
            First name*
          </Text>
          <Input
            type="text"
            placeholder="Enter your first name"
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

        <Button className="bg-green-600 hover:bg-green-700 mt-4 px-6 py-3 text-white">
          Join newsletter
        </Button>

        <Text className="text-gray-600 mt-2 text-sm">
          *NB* We will not spam your inbox, and you can also unsubscribe at any
          time.
        </Text>
      </Box>
    </Box>
  );
};
