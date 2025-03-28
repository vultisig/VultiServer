import { get } from "@/modules/core/services/httpService";
import { PluginMap } from "../models/marketplace";

const getMarketplaceUrl = () => import.meta.env.VITE_MARKETPLACE_URL;

const MarketplaceService = {
  /**
   * Get plugins from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched plugins.
   */
  getPlugins: async (skip: number, take: number): Promise<PluginMap> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/plugins?skip=${skip}&take=${take}`;
      const plugins = await get(endpoint);
      return plugins;
    } catch (error) {
      console.error("Error getting plugins:", error);
      throw error;
    }
  },
};

export default MarketplaceService;
