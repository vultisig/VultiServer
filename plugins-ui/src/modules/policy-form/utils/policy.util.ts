import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key:
      "037c4092a5928c32722df3ac6d02955b20a868fcc58d03ef5e38b1589e09349257", // TODO: get Vault's pub key
    plugin_type,
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
