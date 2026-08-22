import { Flex, Text } from "ui";
import { NotFoundIllustration } from "@/components/ui/illustrations/empty-state-illustrations";

export const PublicPortalNotFoundState = ({
  description,
  title,
}: {
  description: string;
  title: string;
}) => (
  <Flex
    align="center"
    className="bg-background min-h-dvh px-6"
    justify="center"
  >
    <Flex align="center" className="max-w-md text-center" direction="column">
      <NotFoundIllustration className="w-52" />
      <Text as="h1" className="mt-5 text-2xl font-semibold">
        {title}
      </Text>
      <Text className="mt-2" color="muted">
        {description}
      </Text>
    </Flex>
  </Flex>
);
