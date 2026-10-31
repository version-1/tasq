import styles from "./index.module.css";

type SwitchProps = {
  checked?: boolean;
  disabled?: boolean;
} & (
  | { decorative: true; checked: boolean; "aria-label"?: never; onChange?: never }
  | { decorative?: false; "aria-label": string; onChange?: (checked: boolean) => void }
);

export function Switch({
  "aria-label": ariaLabel,
  checked,
  decorative = false,
  disabled,
  onChange,
}: SwitchProps) {
  return (
    <span
      aria-hidden={decorative ? true : undefined}
      className={styles.switch}
      data-checked={decorative ? checked : undefined}
      data-disabled={decorative ? disabled : undefined}
    >
      {!decorative ? <input
        aria-label={ariaLabel}
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange?.(event.target.checked)}
        type="checkbox"
      /> : null}
      <span className={styles.track} aria-hidden="true">
        <span className={styles.thumb} />
      </span>
    </span>
  );
}
