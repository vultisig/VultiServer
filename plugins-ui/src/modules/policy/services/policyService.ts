import { post, get, put, remove } from "@/modules/core/services/httpService";
import { PluginPolicy, PolicyTransactionHistory } from "../models/policy";

const PUBLIC_KEY = import.meta.env.VITE_PUBLIC_KEY;

const PolicyService = {
  /**
   * Posts a new policy to the API.
   * @param {PluginPolicy} pluginPolicy - The policy to be created.
   * @returns {Promise<Object>} A promise that resolves to the created policy.
   */
  createPolicy: async (pluginPolicy: PluginPolicy): Promise<PluginPolicy> => {
    try {
      const endpoint = "/plugin/policy";
      const newPolicy = await post(endpoint, pluginPolicy);
      return newPolicy;
    } catch (error) {
      console.error("Error creating policy:", error);
      throw error;
    }
  },

  /**
   * Updates policy to the API.
   * @param {PluginPolicy} pluginPolicy - The policy to be created.
   * @returns {Promise<Object>} A promise that resolves to the created policy.
   */
  updatePolicy: async (pluginPolicy: PluginPolicy): Promise<PluginPolicy> => {
    try {
      const endpoint = "/plugin/policy";
      const newPolicy = await put(endpoint, pluginPolicy);
      return newPolicy;
    } catch (error) {
      console.error("Error updating policy:", error);
      throw error;
    }
  },

  /**
   * Get policies from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched policies.
   */
  getPolicies: async (): Promise<PluginPolicy[]> => {
    try {
      const endpoint = "/plugin/policy";
      const newPolicy = await get(endpoint, {
        headers: {
          plugin_type: "dca", // todo remove hardcoding once we have the marketplace
          public_key: PUBLIC_KEY,
        },
      });
      return newPolicy;
    } catch (error) {
      console.error("Error getting policies:", error);
      throw error;
    }
  },

  /**
   * Get policy transaction history from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched policies.
   */
  getPolicyTransactionHistory: async (
    policyId: string
  ): Promise<PolicyTransactionHistory[]> => {
    try {
      const endpoint = `/plugin/policy/history/${policyId}`;
      const history = await get(endpoint, {
        headers: {
          public_key: PUBLIC_KEY,
        },
      });
      return history;
    } catch (error) {
      console.error("Error getting policy history:", error);

      throw error;
    }
  },

  /**
   * Delete policy from the API.
   * @param {id} string - The policy to be deleted.
   */
  deletePolicy: async (id: string) => {
    try {
      const endpoint = `/plugin/policy/${id}`;
      return await remove(endpoint);
    } catch (error) {
      console.error("Error deleting policy:", error);
      throw error;
    }
  },
};

export default PolicyService;
