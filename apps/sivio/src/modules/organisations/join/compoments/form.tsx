import { Box, Button, Text } from "ui";
import { Container } from "@/components/ui";

export const Form = () => {
  return (
    <Container className="max-w-4xl py-10">
      <Text className="mb-10 text-4xl font-semibold">
        Application to join the Africa Giving platform
      </Text>

      <form>
        <Button className="mx-auto" size="lg">
          Submit Application
        </Button>
      </form>
    </Container>
  );
};
