import styles from "./index.module.css";

export function Pagination({
  nextDisabled,
  nextLabel,
  onNext,
  onPrevious,
  previousDisabled,
  previousLabel,
  summary,
}: {
  nextDisabled: boolean;
  nextLabel: string;
  onNext: () => void;
  onPrevious: () => void;
  previousDisabled: boolean;
  previousLabel: string;
  summary: string;
}) {
  return (
    <nav className={styles.pagination} aria-label={summary}>
      <button type="button" disabled={previousDisabled} onClick={onPrevious}>
        {previousLabel}
      </button>
      <span>{summary}</span>
      <button type="button" disabled={nextDisabled} onClick={onNext}>
        {nextLabel}
      </button>
    </nav>
  );
}
