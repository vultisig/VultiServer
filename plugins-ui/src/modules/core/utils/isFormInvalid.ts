import { GlobalError } from "react-hook-form"

export const isFormInvalid = (err: { error?: GlobalError }) => {
  if (Object.keys(err).length > 0) return true
  return false
}