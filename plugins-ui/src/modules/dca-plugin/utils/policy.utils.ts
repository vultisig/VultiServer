import { ALLOCATE_TOKEN, BUY_TOKEN } from "../data/tokens";
import { PluginFormData, Policy } from "../models/policy";

export const generatePolicy = (
  submitData: PluginFormData,
  data?: Policy
): Policy => {
  if (data) {
    data.policy = {
      chain_id: "1", // hardcoded for now
      source_token_id: submitData.allocateAsset,
      destination_token_id: submitData.buyAsset,
      total_amount: submitData.amount,
      total_orders: submitData.orders,
      schedule: {
        frequency: submitData.frequency,
        interval: submitData.interval,
        start_time: Date.now().toString(),
      },
    };
    return data;
  }

  return {
    id: "",
    public_key: "8540b779a209ef961bf20618b8e22c678e7bfbad37ec0",
    plugin_type: "dca",
    policy: {
      chain_id: "1", // hardcoded for now
      source_token_id: submitData.allocateAsset,
      destination_token_id: submitData.buyAsset,
      total_amount: submitData.amount,
      total_orders: submitData.orders,
      schedule: {
        frequency: submitData.frequency,
        interval: submitData.interval,
        start_time: Date.now().toString(),
      },
    },
  };
};

export const setDefaultFormValues = (data?: Policy): PluginFormData => {
  const defaultValues: PluginFormData = {
    allocateAsset: ALLOCATE_TOKEN,
    buyAsset: BUY_TOKEN,
    orders: "",
    amount: "0",
    interval: "",
    frequency: "minute",
  };

  if (data) {
    defaultValues.allocateAsset = data.policy.source_token_id;
    defaultValues.buyAsset = data.policy.destination_token_id;
    defaultValues.amount = data.policy.total_amount;
    defaultValues.orders = data.policy.total_orders;
    defaultValues.interval = data.policy.schedule.interval;
    defaultValues.frequency = data.policy.schedule.frequency;
  }

  return defaultValues;
};
