import { ArrowLeftIcon } from "icons";
import { Box, Button, Text } from "ui";
import { ComplexusLogo } from "@/components/ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <ComplexusLogo className="h-8 text-white" />
        <Text className="mb-6 mt-10" fontSize="3xl">
          There was an error
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Oops! something went wrong. Please try again.
        </Text>
        <Button
          className="gap-1 pl-2"
          color="tertiary"
          leftIcon={<ArrowLeftIcon className="h-[1.05rem] w-auto" />}
          onClick={() => {
            if (typeof window !== "undefined") {
              window.location.reload();
            }
          }}
        >
          Reload page
        </Button>
      </Box>
    </Box>
  );
}
