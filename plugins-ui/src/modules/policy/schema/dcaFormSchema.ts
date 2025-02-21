import {
  USDC_TOKEN,
  WETH_TOKEN,
  supportedTokens,
} from "@/modules/shared/data/tokens";
import { asNumber, RJSFSchema, UiSchema } from "@rjsf/utils";
import { Policy } from "../models/policy";
import { SummaryData } from "@/modules/shared/summary/summary.model";

export const schema: RJSFSchema = {
  title: "DCA Plugin Policy",
  type: "object",
  required: [
    "total_amount",
    "source_token_id",
    "destination_token_id",
    "total_orders",
  ],
  properties: {
    chain_id: {
      type: "string",
    },
    total_amount: {
      title: "I want to Allocate",
      type: "number",
    },
    source_token_id: {
      type: "string",
    },
    destination_token_id: {
      title: "I want to buy",
      type: "string",
    },
    schedule: {
      type: "object",
      items: {
        type: "object",
      },
      required: ["interval", "frequency"],
      properties: {
        interval: {
          title: "Every",
          type: "number",
          minimum: 1,
        },
        frequency: {
          type: "string",
          title: "Time",
          enum: ["5-minutely", "hourly", "daily", "weekly", "monthly"],
          default: "5-minutely",
        },
      },
    },
    total_orders: {
      title: "Over (orders)",
      type: "number",
      minimum: 1,
    },
    price_range: {
      type: "object",
      items: {
        type: "object",
        required: ["title"],
      },
      properties: {
        min: {
          title: "Price Range (optional)",
          type: "number",
        },
        max: {
          type: "number",
        },
      },
    },
  },
};

const uiSchema: UiSchema = {
  chain_id: { "ui:widget": "hidden" },
  total_amount: {
    "ui:classNames": "input-background stacked-input",
    "ui:style": {
      display: "inline-block",
      width: "48%",
      marginRight: "2%",
      boxSizing: "border-box",
      verticalAlign: "top",
    },
  },
  source_token_id: {
    "ui:widget": "TokenSelector",
    "ui:options": {
      label: false,
      classNames: "input-background stacked-input",
    },
    "ui:style": {
      display: "inline-block",
      width: "48%",
      boxSizing: "border-box",
      verticalAlign: "top",
      marginTop: "37px",
    },
  },
  destination_token_id: {
    "ui:widget": "TokenSelector",
    "ui:options": {
      classNames: "input-background stacked-input",
    },
  },
  schedule: {
    "ui:options": { label: false },
    "ui:classNames": "form-row",
    frequency: {
      "ui:classNames": "input-background stacked-input",
      "ui:style": {
        display: "flex",
        flexDirection: "column",
      },
    },
    interval: {
      "ui:classNames": "input-background stacked-input",
      "ui:style": {
        display: "flex",
        flexDirection: "column",
      },
    },
  },
  total_orders: {
    "ui:classNames": "input-background stacked-input",
  },
  price_range: {
    "ui:options": { label: false, classNames: "form-row" },
    min: {
      "ui:options": {
        classNames: "input-background stacked-input",
        placeholder: "Min Price",
      },
      "ui:style": {
        display: "flex",
        flexDirection: "column",
        justifyContent: "flex-end",
      },
    },
    max: {
      "ui:options": {
        classNames: "input-background stacked-input",
        label: false,
        placeholder: "Max Price",
      },
      "ui:style": {
        display: "flex",
        flexDirection: "column",
        justifyContent: "flex-end",
      },
    },
  },
};

export const getUiSchema = (policyId: string, policy: Policy): UiSchema => {
  if (policyId) {
    return {
      ...{
        "ui:options": {
          source_token_id: policy.source_token_id,
          destination_token_id: policy.destination_token_id,
        },
      },
      ...{
        "ui:description": "Edit configuration settings for this pair.",
      },
      ...uiSchema,
      ...{
        "ui:submitButtonOptions": {
          submitText: "Edit policy",
        },
      },
    };
  }
  return {
    ...{
      "ui:description": "Set up configuration settings for DCA Plugin Policy",
    },
    ...uiSchema,
    ...{
      "ui:submitButtonOptions": {
        submitText: "Create new policy",
      },
    },
  };
};

export const defaultFormData = {
  chain_id: "1",
  source_token_id: WETH_TOKEN,
  destination_token_id: USDC_TOKEN,
  total_amount: asNumber(null),
  total_orders: asNumber(null),
  schedule: {
    frequency: "minute",
    interval: asNumber(null),
    start_time: new Date().toISOString(),
  },
};

export const getFormData = (data: Policy): Policy => {
  let formData = {
    chain_id: data?.chain_id,
    source_token_id: data?.source_token_id,
    destination_token_id: data?.destination_token_id,
    total_amount: asNumber(data?.total_amount as string),
    total_orders: asNumber(data?.total_orders as string),
    schedule: {
      frequency: (data?.schedule as Policy)?.frequency,
      interval: asNumber((data?.schedule as Policy)?.interval as string),
      start_time: (data?.schedule as Policy)?.start_time,
    },
  };

  if (data?.price_range) {
    formData = {
      ...formData,
      ...{
        price_range: {
          min: asNumber((data?.price_range as Policy)?.min as string),
          max: asNumber((data?.price_range as Policy)?.max as string),
        },
      },
    };
  }

  return formData;
};

export const getSummaryData = (
  formData: Record<string, unknown>
): SummaryData | null => {
  return {
    title: "DCA Summary",
    data: [
      {
        key: "Sell Total",
        value: `${formData.total_amount} ${supportedTokens[formData.source_token_id as string].name}`,
      },
      {
        key: "Sell per order",
        value: `${(formData.total_amount as number) / (formData.total_orders as number)} ${supportedTokens[formData.source_token_id as string].name}`,
      },
      {
        key: "To buy",
        value: `${supportedTokens[formData.destination_token_id as string].name}`,
      },
      { key: "Platform fee", value: "0.1%" },
    ],
  };
};
