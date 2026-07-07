import type { ReactNode } from "react";
import { BreadCrumbs, Box, Flex, Text } from "ui";

export const PageHeader = ({
  actions,
  description,
  eyebrow,
  icon,
  parentHref = "/overview",
  titleActions,
  title,
}: {
  actions?: ReactNode;
  description?: string;
  eyebrow?: string;
  icon?: ReactNode;
  parentHref?: string;
  titleActions?: ReactNode;
  title: string;
}) => {
  return (
    <Box>
      <Flex
        align="center"
        className="border-border/80 h-16 gap-4 border-b-[0.5px] px-5 md:px-7"
        justify="between"
      >
        <BreadCrumbs
          breadCrumbs={[
            ...(eyebrow
              ? [
                  {
                    name: eyebrow,
                    url: parentHref,
                  },
                ]
              : []),
            {
              name: title,
              icon,
            },
          ]}
          className="min-w-0"
        />
        {actions ? <Box className="shrink-0">{actions}</Box> : null}
      </Flex>

      {description ? (
        <Box className="px-5 pt-4 pb-5 md:px-7">
          <Flex align="start" className="gap-4" justify="between">
            <Box className="min-w-0">
              <Text
                as="h1"
                className="text-[1.55rem] leading-tight"
                fontWeight="semibold"
              >
                {title}
              </Text>
              <Text className="mt-1 max-w-3xl text-[0.98rem]" color="muted">
                {description}
              </Text>
            </Box>
            {titleActions ? (
              <Box className="shrink-0 pt-0.5">{titleActions}</Box>
            ) : null}
          </Flex>
        </Box>
      ) : null}
    </Box>
  );
};
