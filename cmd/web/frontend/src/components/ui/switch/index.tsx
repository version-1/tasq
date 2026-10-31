import styles from "./index.module.css";

type SwitchProps = {
  "aria-label": string;
  checked?: boolean;
  disabled?: boolean;
  onChange?: (checked: boolean) => void;
};

export function Switch({
  "aria-label": ariaLabel,
  checked,
  disabled,
  onChange,
}: SwitchProps) {
  return (
    <span className={styles.switch}>
      <input
        aria-label={ariaLabel}
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange?.(event.target.checked)}
        type="checkbox"
      />
      <span className={styles.track} aria-hidden="true">
        <span className={styles.thumb} />
      </span>
    </span>
  );
}
