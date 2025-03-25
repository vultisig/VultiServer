import { get, post, put, remove } from "@/modules/core/services/httpService";
import {
  PluginPolicy,
  PolicySchema,
  PolicyTransactionHistory,
} from "@/modules/policy/models/policy";
import PolicyService from "@/modules/policy/services/policyService";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, it, expect, vi, afterEach, Mock, beforeEach } from "vitest";

vi.mock("@/modules/core/services/httpService", () => ({
  post: vi.fn(),
  put: vi.fn(),
  get: vi.fn(),
  remove: vi.fn(),
}));

describe("PolicyService", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_PLUGIN_URL", "https://mock-api.com");
    localStorage.setItem("publicKey", "publicKey");
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
    localStorage.clear();
  });

  describe("createPolicy", () => {
    it("should call /plugin/policy endpoint and return json object", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockResponse: PluginPolicy = {
        id: "1",
        public_key: "public_key",
        plugin_type: "pluginType",
        is_ecdsa: true,
        chain_code_hex: "",
        derive_path: "",
        plugin_id: "",
        plugin_version: "0.0.1",
        policy_version: "0.0.1",
        active: true,
        signature: "signature",
        policy: {},
      };

      (post as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.createPolicy(mockPolicy);

      expect(post).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy",
        mockPolicy
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when post fails", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockError = new Error("API Error");

      (post as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(PolicyService.createPolicy(mockPolicy)).rejects.toThrow(
        "API Error"
      );

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error creating policy:",
        mockError
      );
    });
  });

  describe("updatePolicy", () => {
    it("should call /plugin/policy endpoint and return json object", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockResponse: PluginPolicy = {
        id: "1",
        public_key: "public_key",
        plugin_type: "pluginType",
        is_ecdsa: true,
        chain_code_hex: "",
        derive_path: "",
        plugin_id: "",
        plugin_version: "0.0.1",
        policy_version: "0.0.1",
        active: true,
        signature: "signature",
        policy: {},
      };

      (put as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.updatePolicy(mockPolicy);

      expect(put).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy",
        mockPolicy
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when put fails", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockError = new Error("API Error");

      (put as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(PolicyService.updatePolicy(mockPolicy)).rejects.toThrow(
        "API Error"
      );

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error updating policy:",
        mockError
      );
    });
  });

  describe("getPolicies", () => {
    it("should call /plugin/policy endpoint and return json object", async () => {
      const PUBLIC_KEY = localStorage.getItem("publicKey");
      const mockRequest = {
        headers: {
          Authorization: "Bearer null",
          plugin_type: "pluginType",
          public_key: PUBLIC_KEY,
        },
      };
      const mockResponse: PluginPolicy[] = [
        {
          id: "1",
          public_key: "public_key",
          plugin_type: "pluginType",
          active: true,
          is_ecdsa: true,
          chain_code_hex: "",
          derive_path: "",
          plugin_id: "",
          plugin_version: "0.0.1",
          policy_version: "0.0.1",
          signature: "signature",
          policy: {},
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.getPolicies("pluginType");

      expect(get).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(PolicyService.getPolicies("pluginType")).rejects.toThrow(
        "API Error"
      );

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policies:",
        mockError
      );
    });
  });

  describe("getPolicyTransactionHistory", () => {
    it("should call /plugin/policy/history/{policyId} endpoint and return json object", async () => {
      const PUBLIC_KEY = localStorage.getItem("publicKey");
      console.log(111, PUBLIC_KEY);

      const mockRequest = {
        headers: {
          Authorization: "Bearer null",
          public_key: PUBLIC_KEY,
        },
      };
      const mockResponse: PolicyTransactionHistory[] = [
        {
          id: "1",
          updated_at: "03/07/25",
          status: "MINED",
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result =
        await PolicyService.getPolicyTransactionHistory("policyId");

      expect(get).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy/history/policyId",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        PolicyService.getPolicyTransactionHistory("policyId")
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policy history:",
        mockError
      );
    });
  });

  describe("deletePolicy", () => {
    it("should call /plugin/policy/{policyId} endpoint and return nothing", async () => {
      (remove as Mock).mockResolvedValue(undefined);

      const result = await PolicyService.deletePolicy("policyId", "signature");

      expect(remove).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy/policyId",
        {
          signature: "signature",
        }
      );
      expect(result).toEqual(undefined);
    });

    it("throws an error when remove fails", async () => {
      const mockError = new Error("API Error");

      (remove as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        PolicyService.deletePolicy("policyId", "signature")
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error deleting policy:",
        mockError
      );
    });
  });

  describe("getPolicySchema", () => {
    it("should call /plugins/schema endpoint and return json object", async () => {
      const mockRequest = {
        headers: {
          plugin_type: "pluginType",
        },
      };
      const mockResponse: PolicySchema[] = [
        {
          form: {
            schema: {},
            uiSchema: {},
            plugin_version: "",
            policy_version: "",
            plugin_type: "",
          },
          table: {
            columns: [],
            mapping: {},
          },
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.getPolicySchema("pluginType");

      expect(get).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy/schema",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(PolicyService.getPolicySchema("pluginType")).rejects.toThrow(
        "API Error"
      );

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policy schema:",
        mockError
      );
    });
  });
});
