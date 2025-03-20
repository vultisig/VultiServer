import { Policy } from "@/modules/policy/models/policy";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, expect, it } from "vitest";

describe("generatePolicy", () => {
  it("should render Pause/Pause button for each policy depending on its state", () => {
    const policyData: Policy = {
      someNumberInput: 5,
      someBooleanInput: false,
      someStringInput: "text",
      someNestedInput: {
        someNumberInput: 5,
        someBooleanInput: false,
        someStringInput: "text",
      },
      someNullInput: null,
      someUndefinedInput: undefined,
    };

    const result = generatePolicy(
      "0.0.1",
      "0.0.1",
      "pluginType",
      "",
      policyData
    );

    expect(result).toStrictEqual({
      active: true,
      chain_code_hex: "",
      derive_path: "",
      id: "",
      is_ecdsa: true,
      plugin_id: "TODO",
      plugin_type: "pluginType",
      plugin_version: "0.0.1",
      policy: {
        someNumberInput: "5",
        someBooleanInput: "false",
        someStringInput: "text",
        someNestedInput: {
          someNumberInput: "5",
          someBooleanInput: "false",
          someStringInput: "text",
        },
        someNullInput: "null",
        someUndefinedInput: "undefined",
      },
      policy_version: "0.0.1",
      public_key: "",
      signature: "",
    });
  });
});
