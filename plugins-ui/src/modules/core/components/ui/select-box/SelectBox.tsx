type SelectBoxProps = {
  name: string;
  options: string[];
  defaultValue: string;
};
export default function SelectBox({ options, defaultValue }: SelectBoxProps) {
  return (
    <>
      <select aria-label="select option" defaultValue={defaultValue}>
        {options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    </>
  );
}
