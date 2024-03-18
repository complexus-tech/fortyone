"use client";
import { Box, Button, Container, Flex, Text } from "ui";
import { ArrowDownIcon } from "icons";
import { StoryStatusIcon } from "@/components/ui";
import { BodyContainer } from "../shared";
import type { Project } from "./project";
import { ProjectCard } from "./project";
import { ProjectsHeader } from "./header";

export const ProjectsList = ({ projects }: { projects: Project[] }) => {
  return (
    <>
      <ProjectsHeader />
      <BodyContainer>
        <Container className="sticky top-0 z-[1] select-none bg-gray-50 py-2 backdrop-blur dark:bg-dark-200/60">
          <Flex align="center" justify="between">
            <Flex align="center" gap={2}>
              <StoryStatusIcon />
              <Text fontWeight="medium">In Progress</Text>
              <Text color="muted">3</Text>
            </Flex>
            <Flex align="center" gap={5}>
              <Text className="w-32 text-left" color="muted">
                Progress
              </Text>
              <Text className="w-32 text-left" color="muted">
                Start date
              </Text>
              <Text className="w-32 text-left" color="muted">
                Target
              </Text>
              <Text className="w-12 text-left" color="muted">
                Lead
              </Text>
              <Text className="w-28 text-left" color="muted">
                Created
              </Text>
              <Box className="w-8">
                <Button
                  className="aspect-square"
                  color="tertiary"
                  rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
                  size="sm"
                  variant="outline"
                >
                  <span className="sr-only">Collapse</span>
                </Button>
              </Box>
            </Flex>
          </Flex>
        </Container>
        {projects.map(({ id, name }) => (
          <ProjectCard key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
};
