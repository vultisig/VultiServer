import { Policy } from "@/modules/policy/models/policy";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, expect, it } from "vitest";

const PUBLIC_KEY = import.meta.env.VITE_PUBLIC_KEY;

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

    const result = generatePolicy("pluginType", "", policyData);

    expect(result).toStrictEqual({
      active: true,
      id: "",
      plugin_type: "pluginType",
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
      public_key: PUBLIC_KEY,
      signature: "",
    });
  });
});
