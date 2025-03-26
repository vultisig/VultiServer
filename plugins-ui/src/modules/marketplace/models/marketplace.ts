export type ViewFilter = "grid" | "list";

type Plugin = {
  id: string;
  type: string;
  title: string;
  description: string;
  metadata: {};
  server_endpoint: string;
  pricing_id: string;
};

export type PluginMap = {
  plugins: Plugin[];
  total_count: number;
};
