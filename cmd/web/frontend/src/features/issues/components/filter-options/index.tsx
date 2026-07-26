import { useEffect, useRef, useState } from "react";
import { IconProxy } from "@/components/ui/icon-proxy";
import styles from "./index.module.css";

type FilterOption<T extends string | number> = {
  children?: FilterOption<T>[];
  label: string;
  value?: T;
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
  const selectedOptions = selectableOptions(options).filter((option) => (
    option.value !== undefined && selectedValues.includes(option.value)
  ));
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

  function handleToggle(values: FilterValue[]) {
    setDraftValues((current) => (
      values.every((value) => current.includes(value))
        ? current.filter((item) => !values.includes(item))
        : [...new Set([...current, ...values])]
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
                <FilterOptionRow
                  draftValues={draftValues}
                  key={option.label}
                  isNested={false}
                  onToggle={handleToggle}
                  option={option}
                />
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

function FilterOptionRow({
  draftValues,
  isNested,
  onToggle,
  option,
}: {
  draftValues: FilterValue[];
  isNested: boolean;
  onToggle: (values: FilterValue[]) => void;
  option: FilterOption<FilterValue>;
}) {
  const checkboxRef = useRef<HTMLInputElement>(null);
  const values = selectableOptions([option]).flatMap((item) => item.value === undefined ? [] : [item.value]);
  const selectedCount = values.filter((value) => draftValues.includes(value)).length;
  const isChecked = values.length > 0 && selectedCount === values.length;
  const isIndeterminate = selectedCount > 0 && !isChecked;

  useEffect(() => {
    if (checkboxRef.current) {
      checkboxRef.current.indeterminate = isIndeterminate;
    }
  }, [isIndeterminate]);

  return (
    <div className={option.children?.length ? styles.optionGroup : undefined}>
      <label className={[
        styles.optionRow,
        option.children?.length ? styles.groupRow : "",
        isNested ? styles.nestedOptionRow : "",
      ].join(" ")}>
        <input
          checked={isChecked}
          ref={checkboxRef}
          type="checkbox"
          onChange={() => onToggle(values)}
        />
        <span>{option.label}</span>
      </label>
      {option.children?.length ? (
        <div className={styles.optionChildren}>
          {option.children.map((child) => (
            <FilterOptionRow
              draftValues={draftValues}
              isNested={true}
              key={child.value ?? child.label}
              onToggle={onToggle}
              option={child}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
}

function selectableOptions(options: FilterOption<FilterValue>[]): FilterOption<FilterValue>[] {
  return options.flatMap((option) => (
    option.children?.length ? selectableOptions(option.children) : [option]
  ));
}
