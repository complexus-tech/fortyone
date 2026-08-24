import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";
import { ProductScreenshot } from "./product-screenshot";

export const HeroProductScreenshot = () => (
  <section aria-label="FortyOne Kanban board">
    <ProductScreenshot
      alt="FortyOne My Work Kanban board showing prioritized work moving through delivery"
      containerClassName="mt-8 md:mt-10"
      cropBrowserOnMobile
      darkImage={myWorkBoardDark}
      lightImage={myWorkBoardLight}
      priority
      reveal={false}
      url="https://fortyone.app/my-work"
    />
  </section>
);
