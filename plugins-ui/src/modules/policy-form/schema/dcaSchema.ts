import {
  ALLOCATE_TOKEN,
  BUY_TOKEN,
  supportedTokens,
} from "@/modules/dca-plugin/data/tokens";
import TokenSelector from "@/modules/shared/token-selector/TokenSelector";
import { asNumber, RJSFSchema, UiSchema } from "@rjsf/utils";
import { Policy } from "../models/policy";
import { SummaryData } from "@/modules/shared/summary/summary.model";

export const schema: RJSFSchema = {
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
      title: "To Buy",
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
      title: "Over",
      type: "number",
      minimum: 1,
    },
    price_range: {
      title: "Price Range (optional)",
      type: "object",
      items: {
        type: "object",
        required: ["title"],
      },
      properties: {
        min: {
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
      verticalAlign: "bottom",
    },
  },
  source_token_id: {
    "ui:widget": TokenSelector,
    "ui:options": {
      label: false,
      classNames: "input-background stacked-input",
    },
    "ui:style": {
      display: "inline-block",
      width: "48%",
      boxSizing: "border-box",
      verticalAlign: "bottom",
    },
  },
  destination_token_id: {
    "ui:widget": TokenSelector,
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
    "ui:classNames": "form-row",
    min: {
      "ui:options": {
        classNames: "input-background stacked-input",
        label: false,
        placeholder: "Min Price",
      },
      "ui:style": {
        display: "flex",
        flexDirection: "column",
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
      },
    },
  },
};

export const getUiSchema = (policyId: string, policy: Policy): UiSchema => {
  if (policyId) {
    return {
      ...{
        "ui:title": `${supportedTokens[policy.source_token_id as string].name}/${supportedTokens[policy.destination_token_id as string].name}`,
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
    ...{ "ui:title": "DCA Plugin Policy" },
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
  source_token_id: ALLOCATE_TOKEN,
  destination_token_id: BUY_TOKEN,
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
