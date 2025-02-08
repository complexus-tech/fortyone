"use client";
import { Button, Container, Flex, Text } from "ui";
import { ArrowLeftIcon } from "icons";
import { useSession } from "next-auth/react";
import { ComplexusLogo, Dot } from "@/components/ui";

export const Navigation = () => {
  const { data: session } = useSession();
  const activeWorkspace = session?.activeWorkspace;
  return (
    <Container className="absolute left-0 top-0 w-full py-6">
      <Flex align="center" justify="between">
        {activeWorkspace ? (
          <Button
            className="gap-1.5 pl-2"
            color="tertiary"
            href="/my-work"
            leftIcon={<ArrowLeftIcon />}
            variant="naked"
          >
            Back to <Dot color={activeWorkspace.color} /> {activeWorkspace.name}
          </Button>
        ) : (
          <ComplexusLogo className="h-7" />
        )}
        <Text color="muted">{session?.user?.email}</Text>
      </Flex>
    </Container>
  );
};
