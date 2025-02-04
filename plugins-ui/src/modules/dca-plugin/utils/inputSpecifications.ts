import { Frequency } from "../models/policy";

export const allocate_from_validation = {
  label: "I want to Allocate",
  type: "number",
  id: "amount",
  placeholder: "",
  validation: {
    required: {
      value: true,
      message: "required",
    },
    min: {
      value: 1,
      message: "min amount is 1",
    },
  },
};

const timeValidations = {
  minute: {
    value: 15,
    message: "min 15 minutes",
  },
  hour: {
    value: 1,
    message: "min 1 hours",
  },
  day: {
    value: 1,
    message: "min 1 days",
  },
  week: {
    value: 1,
    message: "min 1 week",
  },
  month: {
    value: 1,
    message: "min 1 month",
  },
};

export const generateMinTimeInputValidation = (frequency: Frequency) => {
  time_period_validation.validation.min = timeValidations[frequency];
  return time_period_validation;
};

const time_period_validation = {
  label: "Every",
  type: "number",
  id: "interval",
  placeholder: "",
  validation: {
    required: {
      value: true,
      message: "required",
    },
    min: {
      value: 15,
      message: "min 15 minutes", // default
    },
  },
};

export const orders_validation = {
  label: "Over",
  type: "number",
  id: "orders",
  placeholder: "",
  validation: {
    required: {
      value: true,
      message: "required",
    },
    min: {
      value: 2,
      message: "Number of Orders cannot be lower than 2",
    },
  },
};
