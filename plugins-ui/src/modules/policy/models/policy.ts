import { RJSFSchema } from "@rjsf/utils";

export type Policy<T = string | number | boolean | null | undefined> = {
  [key: string]: T | Policy<T>;
};

export type PluginPolicy = {
  id: string;
  public_key: string;
  is_ecdsa: boolean;
  chain_code_hex: string;
  derive_path: string;
  plugin_id: string;
  plugin_version: string;
  policy_version: string;
  plugin_type: string;
  signature: string;
  policy: Policy;
  active: boolean;
};

export type PolicyTransactionHistory = {
  id: string;
  updated_at: string;
  status: string;
};

export type PolicySchema = {
  form: {
    schema: RJSFSchema;
    uiSchema: {};
    plugin_version: string;
    policy_version: string;
    plugin_type: string;
  };
  table: {
    columns: [];
    mapping: {};
  };
};
