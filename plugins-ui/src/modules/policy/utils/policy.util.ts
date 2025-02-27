import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key:
      "0278507737147e65e34e78e565041f651b646a379c6e4cbb144fdda9ac4b84ee80", // TODO: get Vault's pub key
    plugin_type,
    active: true,
    policy: convertToStrings(policy),
    signature: "", // todo this should be implemented
  };
};

function convertToStrings<T extends Record<string, any>>(
  obj: T
): Record<string, string> {
  return Object.fromEntries(
    Object.entries(obj).map(([key, value]) => [
      key,
      typeof value === "object" && value !== null
        ? convertToStrings(value)
        : String(value),
    ])
  ) as Record<string, string>;
}
