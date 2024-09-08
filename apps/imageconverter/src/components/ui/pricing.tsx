"use client";
import { Flex, Text, Box, Button } from "ui";
import { motion } from "framer-motion";
import { Container } from "./container";
import { Blur } from "./blur";

export const Pricing = () => {
  return (
    <Box className="relative mb-20 md:mb-40">
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 z-[2] h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
      <Container className="md:pt-16">
        <Flex
          align="center"
          className="mb-8 mt-12 text-center md:mt-20"
          direction="column"
        >
          <Button
            className="px-3 text-sm md:text-base"
            color="tertiary"
            rounded="full"
            size="sm"
          >
            Pricing
          </Button>
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            transition={{
              duration: 1,
              delay: 0,
            }}
            viewport={{ once: true, amount: 0.5 }}
            whileInView={{ y: 0, opacity: 1 }}
          >
            <Text
              as="h1"
              className="mx-auto mt-6 h-max max-w-3xl pb-2 font-satoshi text-5xl font-bold md:text-6xl"
              color="gradient"
              fontWeight="medium"
            >
              Convert more, spend less.
            </Text>
            <Text fontSize="xl" className="mx-auto mt-6 max-w-3xl">
              ImageConveta is a free while in beta.
            </Text>
          </motion.div>

          {/* <Box className="mt-6">
            <motion.div
              className="flex gap-1 rounded-[0.6rem] bg-dark-200 p-1"
              initial={{ y: 20, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.3,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              {["monthly", "annual"].map((option) => (
                <Button
                  className={cn("px-2.5 capitalize", {
                    "opacity-80": option !== billing,
                  })}
                  color={option === billing ? "primary" : "tertiary"}
                  key={option}
                  onClick={() => {
                    setBilling(option as Billing);
                  }}
                  size="sm"
                  variant={option === billing ? "solid" : "naked"}
                >
                  {option} Billing
                </Button>
              ))}
            </motion.div>
            <motion.p
              className="mt-3"
              initial={{ y: 20, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.6,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Text as="span" color="primary" fontWeight="semibold">
                Save 25%
              </Text>{" "}
              with annual billing 🎉
            </motion.p>
          </Box> */}
        </Flex>
        {/* 
        <Box className="mx-auto grid max-w-6xl grid-cols-1 gap-8 md:grid-cols-3">
          {packages.map((pkg) => (
            <Package
              billing={billing}
              cta={pkg.cta}
              description={pkg.description}
              features={pkg.features}
              key={pkg.name}
              name={pkg.name}
              overview={pkg.overview}
              price={pkg.price}
              recommended={pkg.recommended}
            />
          ))}
        </Box> */}
      </Container>
    </Box>
  );
};
