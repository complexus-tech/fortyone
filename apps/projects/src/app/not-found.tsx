import { ArrowLeftIcon } from "icons";
import { Box, Button, Text } from "ui";
import { ComplexusLogo } from "@/components/ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <ComplexusLogo className="h-16 w-auto" />
        <Text className="mb-6 mt-10" fontSize="4xl">
          404: Project Detour
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Oops! It seems the project path hit a snag. Our team&lsquo;s on it!
          While we clear the roadblock, why not explore other routes to
          productivity?
        </Text>
        <Button
          className="gap-1 pl-2"
          color="tertiary"
          href="/"
          leftIcon={<ArrowLeftIcon className="h-[1.05rem] w-auto" />}
          variant="outline"
        >
          Back to dashboard
        </Button>
      </Box>
    </Box>
  );
}
