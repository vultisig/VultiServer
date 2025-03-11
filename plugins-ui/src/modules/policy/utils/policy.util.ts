import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key:
      "03506115cd0ce1791f583a9c906c2af336bc5decf0e580fb34bffb57aebdfa7610", // TODO: get Vault's pub key
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
