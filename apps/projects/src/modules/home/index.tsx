import { Box, Container } from "ui";
import { BodyContainer } from "../../components/shared/body";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { Contributions } from "./components/contributions";
import { MyStories } from "./components/my-stories";
import { Activities } from "./components/activities";

export const Dashboard = () => {
  return (
    <>
      <Header />
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
    </>
  );
};
