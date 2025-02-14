export type Policy = {
  [key: string]: string | number | null | undefined | Policy;
};

export type PluginPolicy = {
  id: string;
  public_key: string;
  plugin_type: string;
  policy: Policy;
  signature: string;
};
