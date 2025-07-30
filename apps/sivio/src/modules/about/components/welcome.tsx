import { Box, Text, BlurImage } from "ui";
import { Container } from "@/components/ui";

export const Welcome = () => {
  return (
    <Container className="py-28">
      <Box className="grid grid-cols-1 gap-8 md:grid-cols-3">
        <Box>
          <Text
            as="h2"
            className="mb-4 text-5xl font-black leading-tight text-black"
          >
            Welcome to
            <br />
            AfricaGiving
          </Text>
          <Box className="mb-8 h-2 w-20 bg-black" />
          <Box className="mb-4">
            <BlurImage
              alt="Tendai Murisa - Founder of AfricaGiving"
              className="aspect-[9/9.8]"
              quality={100}
              src="/images/about/5.png"
              imageClassName="bg-white"
            />
          </Box>
          <Text
            className="text-center text-xl font-semibold text-gray antialiased"
            style={{ fontFamily: "cursive" }}
          >
            Founder of AfricaGiving
          </Text>
          <Text
            className="text-center text-xl font-semibold text-gray antialiased"
            style={{ fontFamily: "cursive" }}
          >
            Tendai Murisa
          </Text>
        </Box>
        <Box className="md:col-span-2">
          <Box className="flex max-w-3xl flex-col gap-4">
            <Text className="text-lg">
              It gives me great pleasure to welcome you to AfricaGiving a
              platform born out of a simple yet profound belief: Africa’s future
              will be shaped by Africans, through our own actions, resources,
              and solidarity.
            </Text>
            <Text className="text-lg">
              AfricaGiving is more than a digital tool. It is a commitment to
              amplify the incredible work being done by African non-profits,
              some visible, many hidden, who are quietly transforming
              communities across our continent. From remote villages to bustling
              cities, these organisations are building futures, restoring
              dignity, and advancing justice. What they often lack is visibility
              and access to consistent support. Our aim is to change that.
            </Text>
            <Text className="text-lg">
              AfricaGiving raises the visibility of these African-led
              initiatives and connects them with potential givers, individuals
              across the continent and in the diaspora who are eager to
              contribute to Africa’s transformation. Whether you are passionate
              about education, healthcare, disadvantaged women and girls, or
              youth empowerment, this platform allows you to discover, review,
              and support credible causes that resonate with your values.
            </Text>
            <Text className="text-lg">
              We recognise that many organisations doing meaningful work have
              historically been excluded from traditional funding networks
              simply because they are smaller, newer, or less connected.
              AfricaGiving exists to level that playing field, and in doing so,
              to democratise giving itself.
            </Text>
            <Text className="text-lg">
              Our dream is one we know many of you share: a continent where
              every African lives a life of dignity, equity, and opportunity.
              And we believe that dream is achievable—when each of us plays our
              part.
            </Text>
            <Text className="text-lg">
              Thank you for visiting AfricaGiving. I invite you to explore,
              connect, and give. Together, we can build the Africa we all
              imagine.
            </Text>

            <Text className="text-lg">
              Warm regards,
              <br /> Dr Tendai Murisa <br />
              Founder, AfricaGiving & SIVIO Institute
            </Text>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
