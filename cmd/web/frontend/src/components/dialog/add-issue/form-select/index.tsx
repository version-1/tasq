import { FormField } from "../form-field";

export type FormSelectOption = {
  label: string;
  value: string;
};

export function FormSelect({
  label,
  options,
  placeholder,
  value,
  onValueChange,
}: {
  label: string;
  options: FormSelectOption[];
  placeholder?: FormSelectOption;
  value: string;
  onValueChange: (value: string) => void;
}) {
  return (
    <FormField label={label}>
      <select value={value} onChange={(event) => onValueChange(event.target.value)}>
        {placeholder ? <option value={placeholder.value}>{placeholder.label}</option> : null}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </FormField>
  );
}
