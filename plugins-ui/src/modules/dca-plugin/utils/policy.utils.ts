import { PluginFormData, Policy } from "../models/policy";

export const generatePolicy = (submitData: PluginFormData, data?: Policy): Policy => {

    if (data) {
        data.policy = {
            chain_id: "1", // hardcoded for now
            source_token_id: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
            destination_token_id: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
            total_amount: submitData.amount,
            total_orders: submitData.orders,
            schedule: {
                frequency: submitData.frequency,
                interval: submitData.interval,
                start_time: Date.now().toString(),
            }
        }

        return data
    }

    return {
        public_key: "8540b779a209ef961bf20618b8e22c678e7bfbad37ec0",
        plugin_type: "dca",
        policy: {
            chain_id: "1", // hardcoded for now
            source_token_id: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
            destination_token_id: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
            total_amount: submitData.amount,
            total_orders: submitData.orders,
            schedule: {
                frequency: submitData.frequency,
                interval: submitData.interval,
                start_time: Date.now().toString(),
            }
        },
    }
}