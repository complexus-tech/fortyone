import { ChevronLeft, Frown } from "lucide-react";
import { Box, Button, Text } from "ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <Frown className="h-28 w-auto animate-bounce" strokeWidth={1.5} />
        <Text className="mb-4 mt-6" fontSize="3xl" fontWeight="semibold">
          404: Project Detour
        </Text>
        <Text className="mb-3 max-w-md text-center" color="muted">
          Oops! It seems the project path hit a snag. Our team&lsquo;s on it!
          While we clear the roadblock, why not explore other routes to
          productivity?
        </Text>
        <Button href="/" leftIcon={<ChevronLeft className="h-5 w-auto" />}>
          Back to dashboard
        </Button>
      </Box>
    </Box>
  );
}
