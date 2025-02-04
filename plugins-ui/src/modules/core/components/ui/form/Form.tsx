import {
  DefaultValues,
  FormProvider,
  SubmitHandler,
  useForm,
  UseFormReturn,
} from "react-hook-form";

import "./Form.css";

type FormProps<T extends object> = {
  defaultValues: T;
  onSubmit: SubmitHandler<T>;
  render: (methods: UseFormReturn<T>) => React.ReactNode;
};

// Generic Form Component
const Form = <T extends object>({
  defaultValues,
  onSubmit,
  render,
}: FormProps<T>) => {
  const methods = useForm<T>({
    defaultValues: defaultValues as DefaultValues<T>,
  });

  return (
    <FormProvider {...methods}>
      <form className="form" onSubmit={methods.handleSubmit(onSubmit)}>
        {render(methods)}
      </form>
    </FormProvider>
  );
};

export default Form;
