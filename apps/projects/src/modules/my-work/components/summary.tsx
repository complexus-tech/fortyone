import { Box, Container } from "ui";
import { BodyContainer } from "../../../components/shared/body";
import { Overview } from "./summary/overview";
import { Contributions } from "./summary/contributions";
import { MyStories } from "./summary/my-stories";
import { Activities } from "./summary/activities";

export const Summary = () => {
  return (
    <BodyContainer>
      <Container className="pb-4 pt-3">
        <Overview />
        <Contributions />
        <Box className="my-4 grid min-h-[30rem] grid-cols-2 gap-4">
          <MyStories />
          <Activities />
        </Box>
      </Container>
    </BodyContainer>
  );
};
