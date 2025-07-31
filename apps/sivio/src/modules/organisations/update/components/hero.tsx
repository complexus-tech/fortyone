import { Button, Flex, Text } from "ui";
import { DownloadIcon } from "icons";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container>
      <Flex align="center" className="bg-primary p-16" justify="between">
        <Text className="text-7xl font-bold text-white">Update Profile</Text>
        <Button
          color="white"
          rightIcon={<DownloadIcon className="text-dark" />}
          size="lg"
        >
          Download Template
        </Button>
      </Flex>
    </Container>
  );
};
