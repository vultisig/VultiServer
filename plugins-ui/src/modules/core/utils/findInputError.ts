import { FieldErrors, FieldValues, GlobalError } from "react-hook-form";

export function findInputError(errors: FieldErrors<FieldValues>, name: string): { error?: GlobalError } {

  const filtered = Object.keys(errors)
    .filter(key => key.includes(name))
    .reduce((cur, key) => {
      return Object.assign(cur, { error: errors[key] })
    }, {})

  return filtered
}