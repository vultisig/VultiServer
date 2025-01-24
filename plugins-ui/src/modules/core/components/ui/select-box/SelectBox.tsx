import { useFormContext } from "react-hook-form";

type SelectBoxProps = {
    name: string,
    options: string[],
    defaultValue: string
}
export default function SelectBox({ name, options, defaultValue }: SelectBoxProps) {
    const {
        register,
    } = useFormContext()
    return (
        <>
            <select aria-label="select option" {...register(name)} defaultValue={defaultValue}>
                {options.map((option) => <option key={option} value={option}>{option}</option>)}
            </select>
        </>
    );
}