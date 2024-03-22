import { ObjectivesIcon } from "icons";
import { Button } from "ui";
import styles from "./page.module.css";

export default function Page(): JSX.Element {
  return (
    <main className={styles.main}>
      <ObjectivesIcon />
      docs
      <Button>Click me</Button>
    </main>
  );
}
