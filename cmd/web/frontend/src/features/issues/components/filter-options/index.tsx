import { useEffect, useRef, useState } from "react";
import { IconProxy } from "@/components/ui/icon-proxy";
import styles from "./index.module.css";

type FilterOption<T extends string | number> = {
  label: string;
  value: T;
};

type FilterValue = string | number;

export function IssueFilterOptions({
  allLabel,
  applyLabel,
  cancelLabel,
  clearLabel,
  label,
  onChange,
  options,
  selectedCountLabel,
  selectedValues,
}: {
  allLabel: string;
  applyLabel: string;
  cancelLabel: string;
  clearLabel: string;
  label: string;
  onChange: (values: FilterValue[]) => void;
  options: FilterOption<FilterValue>[];
  selectedCountLabel: (count: number) => string;
  selectedValues: FilterValue[];
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [draftValues, setDraftValues] = useState<FilterValue[]>(selectedValues);
  const rootRef = useRef<HTMLDivElement>(null);
  const selectedOptions = options.filter((option) => selectedValues.includes(option.value));
  const draftSelectedCount = draftValues.length;

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    function handlePointerDown(event: PointerEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    }

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) {
      setDraftValues(selectedValues);
    }
  }, [isOpen, selectedValues]);

  function handleToggle(value: FilterValue) {
    setDraftValues((current) => (
      current.includes(value)
        ? current.filter((item) => item !== value)
        : [...current, value]
    ));
  }

  function handleCancel() {
    setDraftValues(selectedValues);
    setIsOpen(false);
  }

  function handleApply() {
    onChange(draftValues);
    setIsOpen(false);
  }

  return (
    <div ref={rootRef} className={styles.filterOptions} role="group" aria-label={label}>
      <div className={styles.dropdown}>
        <button
          type="button"
          className={styles.summary}
          aria-expanded={isOpen}
          onClick={() => setIsOpen((current) => !current)}
        >
          <span className={styles.summaryLabel}>{label}:</span>
          <span className={styles.summaryContent}>
            {selectedOptions.length === 0 ? (
              <span className={styles.emptyValue}>{allLabel}</span>
            ) : (
              selectedOptions.map((option) => (
                <span key={option.value} className={styles.valueChip}>
                  {option.label}
                </span>
              ))
            )}
          </span>
          <IconProxy className={styles.summaryIcon} name="chevron-down" size={18} strokeWidth={2.3} />
        </button>
        {isOpen ? (
          <div className={styles.popover}>
            <div className={styles.optionHeader}>
              <span>{selectedCountLabel(draftSelectedCount)}</span>
              <button type="button" className={styles.clearButton} onClick={() => setDraftValues([])}>
                {clearLabel}
              </button>
            </div>
            <div className={styles.optionList}>
              {options.map((option) => (
                <label key={option.value} className={styles.optionRow}>
                  <input
                    checked={draftValues.includes(option.value)}
                    type="checkbox"
                    onChange={() => handleToggle(option.value)}
                  />
                  <span>{option.label}</span>
                </label>
              ))}
            </div>
            <div className={styles.optionActions}>
              <button type="button" className={styles.cancelButton} onClick={handleCancel}>
                {cancelLabel}
              </button>
              <button type="button" className={styles.applyButton} onClick={handleApply}>
                {applyLabel}
              </button>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}
