export type Policy<T = string | number | boolean | null | undefined> = {
  [key: string]: T | Policy<T>;
};

export type PluginPolicy = {
  id: string;
  public_key: string;
  plugin_type: string;
  active: boolean;
  policy: Policy;
  signature: string;
};

export type PolicyTransactionHistory = {
  id: string;
  updated_at: string;
  status: string;
};
