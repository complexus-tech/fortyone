"use client";

import { Container, Text } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header } from "./components";

export default function Page(): JSX.Element {
  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="pt-4">
          <Text className="mb-2" fontSize="2xl" fontWeight="medium">
            Web design
          </Text>
          <Text color="muted">
            The quick brown fox jumps over the lazy dog.
          </Text>
        </Container>
      </BodyContainer>
    </>
  );
}
