import { ArrowLeftIcon } from "icons";
import { Box, Button, Text } from "ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <svg
          className="h-28 w-auto"
          fill="none"
          height="24"
          viewBox="0 0 24 24"
          width="24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M2.5 12C2.5 7.52166 2.5 5.28249 3.89124 3.89124C5.28249 2.5 7.52166 2.5 12 2.5C16.4783 2.5 18.7175 2.5 20.1088 3.89124C21.5 5.28249 21.5 7.52166 21.5 12C21.5 16.4783 21.5 18.7175 20.1088 20.1088C18.7175 21.5 16.4783 21.5 12 21.5C7.52166 21.5 5.28249 21.5 3.89124 20.1088C2.5 18.7175 2.5 16.4783 2.5 12Z"
            stroke="currentColor"
            strokeWidth="1"
          />
          <path
            d="M2.5 9H21.5"
            stroke="currentColor"
            strokeLinejoin="round"
            strokeWidth="1"
          />
          <path
            d="M6.99981 6H7.00879"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="1"
          />
          <path
            d="M10.9998 6H11.0088"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="1"
          />
        </svg>
        <Text className="mb-4 mt-6" fontSize="2xl" fontWeight="semibold">
          404: Project Detour
        </Text>
        <Text className="mb-3 max-w-md text-center" color="muted">
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
