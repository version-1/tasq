import styles from "./index.module.css";

type Props = {
  children?: React.ReactNode;
  onClick: () => void;
};

export function Button({ children, onClick }: Props) {
  return (
    <button className={styles.primaryButton} type="button" onClick={onClick}>
      {children}
    </button>
  );
}
