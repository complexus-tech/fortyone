import { Box, Container } from "ui";
import { BodyContainer } from "../shared/body";
import { Header } from "./header";
import { Overview } from "./overview";
import { Contributions } from "./contributions";
import { MyStories } from "./my-stories";
import { Activities } from "./activities";

export const Dashboard = () => {
  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="py-4">
          <Overview />
          <Contributions />
          <Box className="my-4 grid min-h-[30rem] grid-cols-2 gap-4">
            <MyStories />
            <Activities />
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
