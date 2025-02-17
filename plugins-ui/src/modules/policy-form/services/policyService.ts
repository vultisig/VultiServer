import { post, get, put, remove } from "@/modules/core/services/httpService";
import { PluginPolicy } from "../models/policy";

const PolicyService = {
  /**
   * Posts a new policy to the API.
   * @param {PluginPolicy} pluginPolicy - The policy to be created.
   * @returns {Promise<Object>} A promise that resolves to the created policy.
   */
  createPolicy: async (pluginPolicy: PluginPolicy) => {
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
  updatePolicy: async (pluginPolicy: PluginPolicy) => {
    try {
      const endpoint = "/plugin/policy";
      const newPolicy = await put(endpoint, pluginPolicy);
      return newPolicy;
    } catch (error) {
      console.error("Error creating policy:", error);
      throw error;
    }
  },

  /**
   * Get policies from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched policies.
   */
  getPolicies: async () => {
    try {
      const endpoint = "/plugin/policy";
      const newPolicy = await get(endpoint, {
        headers: {
          public_key: "037c4092a5928c32722df3ac6d02955b20a868fcc58d03ef5e38b1589e09349257", // TODO: get Vault's pub key
        },
      });
      return newPolicy;
    } catch (error) {
      console.error("Error getting policies:", error);
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
      return remove(endpoint);
    } catch (error) {
      console.error("Error getting policies:", error);
      throw error;
    }
  },
};

export default PolicyService;
