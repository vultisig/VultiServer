import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_version: string,
  policy_version: string,
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key: "",
    is_ecdsa: true,
    chain_code_hex: "",
    derive_path: "",
    plugin_id: "TODO",
    plugin_version,
    policy_version,
    plugin_type,
    signature: "",
    policy: convertToStrings(policy),
    active: true,
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

const getValueByPath = (obj: Record<string, any>, path: string) =>
  path.split(".").reduce((acc, part) => acc?.[part], obj);

export const mapTableColumnData = (
  value: PluginPolicy,
  mapping: Record<string, any>
) => {
  const obj: Record<string, any> = {};

  for (const [key, paths] of Object.entries(mapping)) {
    if (Array.isArray(paths)) {
      // If it's an array, extract multiple values and store as an array
      obj[key] = paths.map((path) => getValueByPath(value, path));
    } else if (paths.includes(",")) {
      // If it's a concatenated value, extract each and join them
      obj[key] = paths
        .split(",")
        .map((path: any) => getValueByPath(value, path.trim()))
        .join(" ");
    } else if (typeof paths === "string") {
      // If it's a direct string path, extract the value
      obj[key] = getValueByPath(value, paths);
    } else {
      // If it's a static value, assign it directly
      obj[key] = paths;
    }
  }

  return obj;
};
