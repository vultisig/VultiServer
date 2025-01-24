import { Message, RegisterOptions, useFormContext } from "react-hook-form"
import { isFormInvalid } from "@/modules/core/utils/isFormInvalid"
import { findInputError } from "@/modules/core/utils/findInputError"

type InputProps = {
    name: string,
    label: string,
    id: string,
    placeholder: string,
    validation: RegisterOptions,
    type?: string,
    className?: string,
}

export const Input = ({
    name,
    label,
    type,
    id,
    placeholder,
    validation,
    className,
}: InputProps) => {
    const {
        register,
        formState: { errors },
    } = useFormContext()

    const inputError = findInputError(errors, name)
    const isInvalid = isFormInvalid(inputError)

    const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
        // Prevent typing the '-' character
        if (event.key === "-") {
            event.preventDefault();
        }
    };

    const handleOnPaste = (event: React.ClipboardEvent<HTMLInputElement>) => {
        // Prevent pasting the '-' character
        const pastedData = event.clipboardData.getData("Text");

        if (+pastedData < 0) {
            event.preventDefault();
        }

    };

    return (
        <>
            <label htmlFor={id}>
                {label}
            </label>
            {isInvalid && inputError?.error?.message && (
                <InputError
                    message={inputError.error.message}
                />
            )}
            <input
                id={id}
                type={type}
                onKeyDown={handleKeyDown}
                onPaste={handleOnPaste}
                placeholder={placeholder}
                {...register(name, validation)}
            />

        </>
    )
}

const InputError = ({ message }: { message: Message }) => {
    return <div className="error">{message}</div>
}
