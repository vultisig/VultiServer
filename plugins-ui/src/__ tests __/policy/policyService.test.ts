import { get, post, put, remove } from "@/modules/core/services/httpService";
import {
  PluginPolicy,
  PolicyTransactionHistory,
} from "@/modules/policy/models/policy";
import PolicyService from "@/modules/policy/services/policyService";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, it, expect, vi, afterEach, Mock } from "vitest";

const PUBLIC_KEY = import.meta.env.VITE_PUBLIC_KEY;

vi.mock("@/modules/core/services/httpService", () => ({
  post: vi.fn(),
  put: vi.fn(),
  get: vi.fn(),
  remove: vi.fn(),
}));

describe("PolicyService", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("createPolicy", () => {
    it("should call /plugin/policy endpoint and return json object", async () => {
      const mockPolicy: PluginPolicy = generatePolicy("dca", "", {});
      const mockResponse: PluginPolicy = {
        id: "1",
        public_key: "public_key",
        plugin_type: "dca",
        active: true,
        signature: "signature",
        policy: {},
        is_ecdsa: true,
        chain_code_hex: "chain_code_hex",
        derive_path: "derive_path",
        plugin_id: "plugin_id",
        plugin_version: "1",
        policy_version: "1",
      };

      (post as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.createPolicy(mockPolicy);

      expect(post).toHaveBeenCalledWith("/plugin/policy", mockPolicy);
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when post fails", async () => {
      const mockPolicy: PluginPolicy = generatePolicy("dca", "", {});
      const mockError = new Error("API Error");

      (post as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => { });

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
      const mockPolicy: PluginPolicy = generatePolicy("dca", "", {});
      const mockResponse: PluginPolicy = {
        id: "1",
        public_key: "public_key",
        plugin_type: "dca",
        active: true,
        signature: "signature",
        policy: {},
        is_ecdsa: true,
        chain_code_hex: "chain_code_hex",
        derive_path: "derive_path",
        plugin_id: "plugin_id",
        plugin_version: "1",
        policy_version: "1",
      };

      (put as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.updatePolicy(mockPolicy);

      expect(put).toHaveBeenCalledWith("/plugin/policy", mockPolicy);
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when put fails", async () => {
      const mockPolicy: PluginPolicy = generatePolicy("dca", "", {});
      const mockError = new Error("API Error");

      (put as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => { });

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
      const mockRequest = {
        headers: {
          plugin_type: "dca",
          public_key: PUBLIC_KEY,
        },
      };
      const mockResponse: PluginPolicy[] = [
        {
          id: "1",
          public_key: "public_key",
          plugin_type: "dca",
          active: true,
          signature: "signature",
          policy: {},
          is_ecdsa: true,
          chain_code_hex: "chain_code_hex",
          derive_path: "derive_path",
          plugin_id: "plugin_id",
          plugin_version: "1",
          policy_version: "1",
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.getPolicies();

      expect(get).toHaveBeenCalledWith("/plugin/policy", mockRequest);
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => { });

      await expect(PolicyService.getPolicies()).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policies:",
        mockError
      );
    });
  });

  describe("getPolicyTransactionHistory", () => {
    it("should call /plugin/policy/history/{policyId} endpoint and return json object", async () => {
      const mockRequest = {
        headers: {
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
        "/plugin/policy/history/policyId",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => { });

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

      expect(remove).toHaveBeenCalledWith("/plugin/policy/policyId");
      expect(result).toEqual(undefined);
    });

    it("throws an error when remove fails", async () => {
      const mockError = new Error("API Error");

      (remove as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => { });

      await expect(PolicyService.deletePolicy("policyId", "signature")).rejects.toThrow(
        "API Error"
      );

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error deleting policy:",
        mockError
      );
    });
  });
});
