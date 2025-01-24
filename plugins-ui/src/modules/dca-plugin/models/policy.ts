export type Policy = {
    id: string,
    public_key: string,
    plugin_type: "dca",
    policy: {
        chain_id: string,
        source_token_id: string,
        destination_token_id: string,
        total_amount: string,
        total_orders: string,
        schedule: {
            frequency: Frequency,
            interval: string,
            start_time: string
        }
    },
}

export type Frequency = "minute" | "hour" | "day" | "week" | "month";