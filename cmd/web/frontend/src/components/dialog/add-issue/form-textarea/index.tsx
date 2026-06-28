import { FormField } from "../form-field";

export function FormTextarea({
  label,
  placeholder,
  rows,
  value,
  onValueChange,
}: {
  label: string;
  placeholder: string;
  rows: number;
  value: string;
  onValueChange: (value: string) => void;
}) {
  return (
    <FormField label={label}>
      <textarea
        value={value}
        onChange={(event) => onValueChange(event.target.value)}
        placeholder={placeholder}
        rows={rows}
      />
    </FormField>
  );
}
