import { post, get, put, remove } from "@/modules/core/services/httpService";
import { Policy } from "../models/policy";

const DCAService = {
    /**
     * Posts a new policy to the API.
     * @param {Policy} policy - The policy to be created.
     * @returns {Promise<Object>} A promise that resolves to the created policy.
     */
    createPolicy: async (policy: Policy) => {
        try {
            const endpoint = '/plugin/policy';
            const newPolicy = await post(endpoint, policy);
            return newPolicy;
        } catch (error) {
            console.error('Error creating policy:', error);
            throw error;
        }
    },

    /**
     * Updates policy to the API.
     * @param {Policy} policy - The policy to be created.
     * @returns {Promise<Object>} A promise that resolves to the created policy.
    */
    updatePolicy: async (policy: Policy) => {
        try {
            const endpoint = '/plugin/policy';
            const newPolicy = await put(endpoint, policy);
            return newPolicy;
        } catch (error) {
            console.error('Error creating policy:', error);
            throw error;
        }
    },

    /**
     * Get policies from the API.
     * @returns {Promise<Object>} A promise that resolves to the fetched policies.
     */
    getPolicies: async () => {
        try {
            const endpoint = '/plugin/policy';
            const newPolicy = await get(endpoint, { headers: { public_key: "8540b779a209ef961bf20618b8e22c678e7bfbad37ec0" } }); // todo do not hardcode
            return newPolicy;
        } catch (error) {
            console.error('Error getting policies:', error);
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
            console.error('Error getting policies:', error);
            throw error;
        }
    },
};

export default DCAService;
